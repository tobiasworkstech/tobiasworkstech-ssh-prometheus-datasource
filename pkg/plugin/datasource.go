package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/tobiasworkstech/ssh-prometheus-datasource/pkg/ssh"
)

type SSHPrometheusSettings struct {
	// SSH Connection
	SSHHost     string `json:"sshHost"`
	SSHPort     int    `json:"-"` // Parsed manually to handle string/int
	SSHUsername string `json:"sshUsername"`
	AuthMethod  string `json:"authMethod"`

	// Prometheus Connection
	PrometheusURL string `json:"prometheusUrl"`

	// Prometheus Authentication
	PrometheusAuthMethod string `json:"prometheusAuthMethod"`
	PrometheusUsername   string `json:"prometheusUsername"`

	// TLS Settings
	TLSSkipVerify    bool `json:"tlsSkipVerify"`
	TLSWithCACert    bool `json:"tlsWithCACert"`
	TLSWithClientCert bool `json:"tlsWithClientCert"`

	// HTTP Settings
	HTTPMethod            string `json:"httpMethod"`
	CustomQueryParameters string `json:"customQueryParameters"`
	Timeout               int    `json:"timeout"`
}

type Datasource struct {
	settings   SSHPrometheusSettings
	secureData map[string]string
	tunnel     *ssh.Tunnel
	tunnelMu   sync.Mutex
	httpClient *http.Client
}

func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var jsonData SSHPrometheusSettings
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse sshPort manually to handle both string and int
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(settings.JSONData, &rawSettings); err == nil {
		if portVal, ok := rawSettings["sshPort"]; ok {
			switch v := portVal.(type) {
			case float64:
				jsonData.SSHPort = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					jsonData.SSHPort = parsed
				}
			}
		}
	}

	// Set defaults
	if jsonData.SSHPort == 0 {
		jsonData.SSHPort = 22
	}
	if jsonData.PrometheusURL == "" {
		jsonData.PrometheusURL = "http://127.0.0.1:9090"
	}
	if jsonData.HTTPMethod == "" {
		jsonData.HTTPMethod = "GET"
	}
	if jsonData.Timeout == 0 {
		jsonData.Timeout = 30
	}
	if jsonData.PrometheusAuthMethod == "" {
		jsonData.PrometheusAuthMethod = "none"
	}

	secureData := settings.DecryptedSecureJSONData

	// Create TLS config
	tlsConfig, err := createTLSConfig(jsonData, secureData)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	ds := &Datasource{
		settings:   jsonData,
		secureData: secureData,
		httpClient: &http.Client{
			Timeout: time.Duration(jsonData.Timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}

	return ds, nil
}

func createTLSConfig(settings SSHPrometheusSettings, secureData map[string]string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: settings.TLSSkipVerify,
	}

	// Load CA certificate
	if settings.TLSWithCACert {
		caCert := secureData["tlsCACert"]
		if caCert != "" {
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}
	}

	// Load client certificate
	if settings.TLSWithClientCert {
		clientCert := secureData["tlsClientCert"]
		clientKey := secureData["tlsClientKey"]
		if clientCert != "" && clientKey != "" {
			cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return tlsConfig, nil
}

func (d *Datasource) Dispose() {
	d.tunnelMu.Lock()
	defer d.tunnelMu.Unlock()

	if d.tunnel != nil {
		d.tunnel.Close()
		d.tunnel = nil
	}
}

func (d *Datasource) ensureTunnel(ctx context.Context) error {
	d.tunnelMu.Lock()
	defer d.tunnelMu.Unlock()

	if d.tunnel != nil && d.tunnel.IsAlive() {
		return nil
	}

	if d.tunnel != nil {
		d.tunnel.Close()
		d.tunnel = nil
	}

	config := ssh.TunnelConfig{
		SSHHost:     d.settings.SSHHost,
		SSHPort:     d.settings.SSHPort,
		SSHUsername: d.settings.SSHUsername,
		AuthMethod:  d.settings.AuthMethod,
	}

	if d.settings.AuthMethod == "password" {
		config.SSHPassword = d.secureData["sshPassword"]
	} else {
		config.SSHPrivateKey = d.secureData["sshPrivateKey"]
		config.SSHKeyPassphrase = d.secureData["sshKeyPassphrase"]
	}

	promURL, err := url.Parse(d.settings.PrometheusURL)
	if err != nil {
		return fmt.Errorf("invalid prometheus URL: %w", err)
	}

	config.RemoteHost = promURL.Hostname()
	port := promURL.Port()
	if port == "" {
		if promURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	config.RemotePort, _ = strconv.Atoi(port)

	tunnel, err := ssh.NewTunnel(config)
	if err != nil {
		return fmt.Errorf("failed to create SSH tunnel: %w", err)
	}

	d.tunnel = tunnel
	log.DefaultLogger.Info("SSH tunnel established", "host", d.settings.SSHHost)
	return nil
}

func (d *Datasource) getLocalURL() string {
	promURL, _ := url.Parse(d.settings.PrometheusURL)
	scheme := promURL.Scheme
	if scheme == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, d.tunnel.LocalAddr())
}

func (d *Datasource) addPrometheusAuth(req *http.Request) {
	switch d.settings.PrometheusAuthMethod {
	case "basic":
		username := d.settings.PrometheusUsername
		password := d.secureData["prometheusPassword"]
		if username != "" || password != "" {
			req.SetBasicAuth(username, password)
		}
	case "bearer":
		token := d.secureData["prometheusBearerToken"]
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	if err := d.ensureTunnel(ctx); err != nil {
		for _, q := range req.Queries {
			response.Responses[q.RefID] = backend.ErrDataResponse(backend.StatusBadGateway, err.Error())
		}
		return response, nil
	}

	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Expr         string `json:"expr"`
	LegendFormat string `json:"legendFormat"`
	Instant      bool   `json:"instant"`
	Range        bool   `json:"range"`
	Interval     string `json:"interval"`
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("failed to parse query: %v", err))
	}

	if qm.Expr == "" {
		return backend.DataResponse{}
	}

	var endpoint string
	params := url.Values{}
	params.Set("query", qm.Expr)

	// Add custom query parameters
	if d.settings.CustomQueryParameters != "" {
		customParams, err := url.ParseQuery(d.settings.CustomQueryParameters)
		if err == nil {
			for k, v := range customParams {
				for _, val := range v {
					params.Add(k, val)
				}
			}
		}
	}

	if qm.Range && !qm.Instant {
		endpoint = "/api/v1/query_range"
		params.Set("start", strconv.FormatInt(query.TimeRange.From.Unix(), 10))
		params.Set("end", strconv.FormatInt(query.TimeRange.To.Unix(), 10))
		step := d.calculateStep(query.TimeRange.From, query.TimeRange.To, query.MaxDataPoints, qm.Interval)
		params.Set("step", strconv.FormatInt(step, 10))
	} else {
		endpoint = "/api/v1/query"
		params.Set("time", strconv.FormatInt(query.TimeRange.To.Unix(), 10))
	}

	reqURL := fmt.Sprintf("%s%s", d.getLocalURL(), endpoint)

	var httpReq *http.Request
	var err error

	if d.settings.HTTPMethod == "POST" {
		httpReq, err = http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(params.Encode()))
		if err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to create request: %v", err))
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
		httpReq, err = http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to create request: %v", err))
		}
	}

	// Add Prometheus authentication
	d.addPrometheusAuth(httpReq)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadGateway, fmt.Sprintf("prometheus request failed: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to read response: %v", err))
	}

	var promResp prometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to parse prometheus response: %v", err))
	}

	if promResp.Status != "success" {
		return backend.ErrDataResponse(backend.StatusBadRequest, promResp.Error)
	}

	frames := d.transformResponse(promResp, qm.LegendFormat, query.RefID)
	return backend.DataResponse{Frames: frames}
}

type prometheusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   struct {
		ResultType string        `json:"resultType"`
		Result     []interface{} `json:"result"`
	} `json:"data"`
}

func (d *Datasource) calculateStep(from, to time.Time, maxDataPoints int64, interval string) int64 {
	if interval != "" {
		if parsed := parseInterval(interval); parsed > 0 {
			return parsed
		}
	}
	rangeSeconds := to.Unix() - from.Unix()
	step := rangeSeconds / maxDataPoints
	if step < 1 {
		step = 1
	}
	return step
}

func parseInterval(interval string) int64 {
	if len(interval) < 2 {
		return 0
	}
	unit := interval[len(interval)-1]
	valueStr := interval[:len(interval)-1]
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0
	}
	switch unit {
	case 's':
		return value
	case 'm':
		return value * 60
	case 'h':
		return value * 3600
	case 'd':
		return value * 86400
	default:
		return 0
	}
}

func (d *Datasource) transformResponse(resp prometheusResponse, legendFormat, refID string) data.Frames {
	var frames data.Frames

	for _, r := range resp.Data.Result {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		metric, _ := result["metric"].(map[string]interface{})
		labels := make(map[string]string)
		for k, v := range metric {
			if s, ok := v.(string); ok {
				labels[k] = s
			}
		}

		name := formatLegend(labels, legendFormat)
		frame := data.NewFrame(name)
		frame.RefID = refID

		var times []time.Time
		var values []float64

		if resp.Data.ResultType == "matrix" {
			valuesRaw, _ := result["values"].([]interface{})
			for _, v := range valuesRaw {
				point, ok := v.([]interface{})
				if !ok || len(point) != 2 {
					continue
				}
				ts, _ := point[0].(float64)
				val, _ := point[1].(string)
				parsedVal, _ := strconv.ParseFloat(val, 64)
				times = append(times, time.Unix(int64(ts), 0))
				values = append(values, parsedVal)
			}
		} else if resp.Data.ResultType == "vector" {
			valueRaw, _ := result["value"].([]interface{})
			if len(valueRaw) == 2 {
				ts, _ := valueRaw[0].(float64)
				val, _ := valueRaw[1].(string)
				parsedVal, _ := strconv.ParseFloat(val, 64)
				times = append(times, time.Unix(int64(ts), 0))
				values = append(values, parsedVal)
			}
		}

		frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
		valueField := data.NewField("value", labels, values)
		frame.Fields = append(frame.Fields, valueField)
		frames = append(frames, frame)
	}

	return frames
}

func formatLegend(labels map[string]string, format string) string {
	if format == "" {
		var parts []string
		for k, v := range labels {
			parts = append(parts, fmt.Sprintf("%s=%q", k, v))
		}
		if len(parts) == 0 {
			return "value"
		}
		return strings.Join(parts, ", ")
	}

	result := format
	for k, v := range labels {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
		result = strings.ReplaceAll(result, "{{ "+k+" }}", v)
	}
	return result
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if err := d.ensureTunnel(ctx); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to establish SSH tunnel: %s", err.Error()),
		}, nil
	}

	reqURL := fmt.Sprintf("%s/api/v1/query?query=1", d.getLocalURL())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to create request: %s", err.Error()),
		}, nil
	}

	// Add Prometheus authentication for health check
	d.addPrometheusAuth(httpReq)

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to connect to Prometheus: %s", err.Error()),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Prometheus authentication failed (401 Unauthorized)",
		}, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Prometheus access forbidden (403 Forbidden)",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Prometheus returned status %d", resp.StatusCode),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "SSH connection and Prometheus are working",
	}, nil
}

func (d *Datasource) handleTestSSH(ctx context.Context, sender backend.CallResourceResponseSender) error {
	config := ssh.TunnelConfig{
		SSHHost:     d.settings.SSHHost,
		SSHPort:     d.settings.SSHPort,
		SSHUsername: d.settings.SSHUsername,
		AuthMethod:  d.settings.AuthMethod,
		RemoteHost:  "localhost",
		RemotePort:  22,
	}

	if d.settings.AuthMethod == "password" {
		config.SSHPassword = d.secureData["sshPassword"]
	} else {
		config.SSHPrivateKey = d.secureData["sshPrivateKey"]
		config.SSHKeyPassphrase = d.secureData["sshKeyPassphrase"]
	}

	// Test SSH connection only (without creating a tunnel)
	err := ssh.TestConnection(config)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: []byte(fmt.Sprintf(`{"status": "error", "message": "SSH connection failed: %s"}`, err.Error())),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"status": "ok", "message": "SSH connection successful"}`),
	})
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	// Handle SSH-only test endpoint
	if req.Path == "test-ssh" {
		return d.handleTestSSH(ctx, sender)
	}

	if err := d.ensureTunnel(ctx); err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
		})
	}

	path := req.Path
	if len(req.URL) > len(req.Path) {
		path = req.URL
	}
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	targetURL := fmt.Sprintf("%s%s", d.getLocalURL(), path)

	var body io.Reader
	var contentType string

	// Check if the body is JSON and convert to form-encoded for Prometheus API
	if len(req.Body) > 0 && req.Method == "POST" {
		var jsonBody map[string]interface{}
		if err := json.Unmarshal(req.Body, &jsonBody); err == nil {
			// Convert JSON to URL-encoded form data for Prometheus
			formData := url.Values{}
			for k, v := range jsonBody {
				switch val := v.(type) {
				case string:
					formData.Set(k, val)
				case float64:
					formData.Set(k, strconv.FormatFloat(val, 'f', -1, 64))
				case int:
					formData.Set(k, strconv.Itoa(val))
				default:
					formData.Set(k, fmt.Sprintf("%v", val))
				}
			}
			body = strings.NewReader(formData.Encode())
			contentType = "application/x-www-form-urlencoded"
		} else {
			body = strings.NewReader(string(req.Body))
		}
	} else if len(req.Body) > 0 {
		body = strings.NewReader(string(req.Body))
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
		})
	}

	// Set content type if we converted to form-encoded
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Add Prometheus authentication for resource calls
	d.addPrometheusAuth(httpReq)

	for k, v := range req.Headers {
		// Skip Content-Type if we already set it
		if contentType != "" && strings.ToLower(k) == "content-type" {
			continue
		}
		for _, val := range v {
			httpReq.Header.Add(k, val)
		}
	}

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
		})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
		})
	}

	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	return sender.Send(&backend.CallResourceResponse{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    respBody,
	})
}

var _ backend.QueryDataHandler = (*Datasource)(nil)
var _ backend.CheckHealthHandler = (*Datasource)(nil)
var _ backend.CallResourceHandler = (*Datasource)(nil)
var _ instancemgmt.InstanceDisposer = (*Datasource)(nil)
var _ = httpadapter.New
