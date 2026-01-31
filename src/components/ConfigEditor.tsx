import React, { ChangeEvent } from 'react';
import {
  InlineField,
  Input,
  SecretInput,
  RadioButtonGroup,
  TextArea,
  FieldSet,
  Select,
  Switch,
  InlineFieldRow,
} from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import {
  SSHPrometheusDataSourceOptions,
  SSHPrometheusSecureJsonData,
  AuthMethod,
  PrometheusAuthMethod,
} from '../types';

interface Props
  extends DataSourcePluginOptionsEditorProps<SSHPrometheusDataSourceOptions, SSHPrometheusSecureJsonData> {}

const sshAuthMethodOptions = [
  { label: 'Password', value: 'password' as AuthMethod },
  { label: 'SSH Key', value: 'key' as AuthMethod },
];

const prometheusAuthMethodOptions: Array<SelectableValue<PrometheusAuthMethod>> = [
  { label: 'No Authentication', value: 'none', description: 'No authentication required' },
  { label: 'Basic Authentication', value: 'basic', description: 'Username and password' },
  { label: 'Bearer Token', value: 'bearer', description: 'Bearer token authentication' },
];

const httpMethodOptions: Array<SelectableValue<'GET' | 'POST'>> = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
];

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onJsonDataChange = <K extends keyof SSHPrometheusDataSourceOptions>(
    key: K,
    value: SSHPrometheusDataSourceOptions[K]
  ) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        [key]: value,
      },
    });
  };

  const onSecureJsonDataChange = <K extends keyof SSHPrometheusSecureJsonData>(
    key: K,
    value: SSHPrometheusSecureJsonData[K]
  ) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        [key]: value,
      },
    });
  };

  const onResetSecureJsonData = (key: keyof SSHPrometheusSecureJsonData) => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        [key]: false,
      },
      secureJsonData: {
        ...secureJsonData,
        [key]: '',
      },
    });
  };

  const sshAuthMethod = jsonData.authMethod || 'password';
  const prometheusAuthMethod = jsonData.prometheusAuthMethod || 'none';

  return (
    <>
      {/* SSH Connection Section */}
      <FieldSet label="SSH Connection">
        <InlineField label="SSH Host" labelWidth={20} tooltip="Remote hostname or IP address to SSH into">
          <Input
            width={40}
            value={jsonData.sshHost || ''}
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('sshHost', e.target.value)}
            placeholder="example.com"
          />
        </InlineField>

        <InlineField label="SSH Port" labelWidth={20} tooltip="SSH port (default: 22)">
          <Input
            width={10}
            type="number"
            value={jsonData.sshPort || 22}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              onJsonDataChange('sshPort', parseInt(e.target.value, 10) || 22)
            }
            placeholder="22"
          />
        </InlineField>

        <InlineField label="SSH Username" labelWidth={20} tooltip="SSH username for authentication">
          <Input
            width={40}
            value={jsonData.sshUsername || ''}
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('sshUsername', e.target.value)}
            placeholder="root"
          />
        </InlineField>
      </FieldSet>

      {/* SSH Authentication Section */}
      <FieldSet label="SSH Authentication">
        <InlineField label="Auth Method" labelWidth={20}>
          <RadioButtonGroup
            options={sshAuthMethodOptions}
            value={sshAuthMethod}
            onChange={(v) => onJsonDataChange('authMethod', v)}
          />
        </InlineField>

        {sshAuthMethod === 'password' && (
          <InlineField label="SSH Password" labelWidth={20} tooltip="Password for SSH authentication">
            <SecretInput
              width={40}
              isConfigured={secureJsonFields?.sshPassword || false}
              value={secureJsonData?.sshPassword || ''}
              onReset={() => onResetSecureJsonData('sshPassword')}
              onChange={(e: ChangeEvent<HTMLInputElement>) => onSecureJsonDataChange('sshPassword', e.target.value)}
              placeholder="Enter SSH password"
            />
          </InlineField>
        )}

        {sshAuthMethod === 'key' && (
          <>
            <InlineField
              label="SSH Private Key"
              labelWidth={20}
              tooltip="PEM-encoded private key (e.g., contents of ~/.ssh/id_rsa)"
            >
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {secureJsonFields?.sshPrivateKey ? (
                  <Input
                    width={40}
                    value="configured"
                    disabled
                    suffix={
                      <button
                        type="button"
                        className="btn btn-link"
                        onClick={() => onResetSecureJsonData('sshPrivateKey')}
                      >
                        Reset
                      </button>
                    }
                  />
                ) : (
                  <TextArea
                    cols={60}
                    rows={8}
                    value={secureJsonData?.sshPrivateKey || ''}
                    onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
                      onSecureJsonDataChange('sshPrivateKey', e.target.value)
                    }
                    placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                  />
                )}
              </div>
            </InlineField>

            <InlineField
              label="Key Passphrase"
              labelWidth={20}
              tooltip="Passphrase for encrypted private keys (leave empty if key is not encrypted)"
            >
              <SecretInput
                width={40}
                isConfigured={secureJsonFields?.sshKeyPassphrase || false}
                value={secureJsonData?.sshKeyPassphrase || ''}
                onReset={() => onResetSecureJsonData('sshKeyPassphrase')}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  onSecureJsonDataChange('sshKeyPassphrase', e.target.value)
                }
                placeholder="Optional passphrase"
              />
            </InlineField>
          </>
        )}
      </FieldSet>

      {/* Prometheus Connection Section */}
      <FieldSet label="Prometheus Connection">
        <InlineField
          label="Prometheus URL"
          labelWidth={20}
          tooltip="Prometheus URL as accessible from the remote SSH host (e.g., http://127.0.0.1:9090)"
        >
          <Input
            width={40}
            value={jsonData.prometheusUrl || 'http://127.0.0.1:9090'}
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('prometheusUrl', e.target.value)}
            placeholder="http://127.0.0.1:9090"
          />
        </InlineField>

        <InlineField label="HTTP Method" labelWidth={20} tooltip="HTTP method used to query Prometheus">
          <Select
            width={20}
            options={httpMethodOptions}
            value={jsonData.httpMethod || 'GET'}
            onChange={(v) => onJsonDataChange('httpMethod', v.value || 'GET')}
          />
        </InlineField>

        <InlineField
          label="Timeout"
          labelWidth={20}
          tooltip="HTTP request timeout in seconds (default: 30)"
        >
          <Input
            width={10}
            type="number"
            value={jsonData.timeout || 30}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              onJsonDataChange('timeout', parseInt(e.target.value, 10) || 30)
            }
            placeholder="30"
          />
        </InlineField>
      </FieldSet>

      {/* Prometheus Authentication Section */}
      <FieldSet label="Prometheus Authentication">
        <InlineField
          label="Authentication"
          labelWidth={20}
          tooltip="Authentication method for the Prometheus server"
        >
          <Select
            width={30}
            options={prometheusAuthMethodOptions}
            value={prometheusAuthMethod}
            onChange={(v) => onJsonDataChange('prometheusAuthMethod', v.value || 'none')}
          />
        </InlineField>

        {prometheusAuthMethod === 'basic' && (
          <>
            <InlineField label="Username" labelWidth={20} tooltip="Basic auth username">
              <Input
                width={40}
                value={jsonData.prometheusUsername || ''}
                onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('prometheusUsername', e.target.value)}
                placeholder="username"
              />
            </InlineField>

            <InlineField label="Password" labelWidth={20} tooltip="Basic auth password">
              <SecretInput
                width={40}
                isConfigured={secureJsonFields?.prometheusPassword || false}
                value={secureJsonData?.prometheusPassword || ''}
                onReset={() => onResetSecureJsonData('prometheusPassword')}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  onSecureJsonDataChange('prometheusPassword', e.target.value)
                }
                placeholder="password"
              />
            </InlineField>
          </>
        )}

        {prometheusAuthMethod === 'bearer' && (
          <InlineField label="Bearer Token" labelWidth={20} tooltip="Bearer token for authentication">
            <SecretInput
              width={40}
              isConfigured={secureJsonFields?.prometheusBearerToken || false}
              value={secureJsonData?.prometheusBearerToken || ''}
              onReset={() => onResetSecureJsonData('prometheusBearerToken')}
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                onSecureJsonDataChange('prometheusBearerToken', e.target.value)
              }
              placeholder="Enter bearer token"
            />
          </InlineField>
        )}
      </FieldSet>

      {/* TLS Settings Section */}
      <FieldSet label="TLS Settings">
        <InlineFieldRow>
          <InlineField
            label="Skip TLS Verification"
            labelWidth={20}
            tooltip="Skip verification of the Prometheus server's TLS certificate"
          >
            <Switch
              value={jsonData.tlsSkipVerify || false}
              onChange={(e) => onJsonDataChange('tlsSkipVerify', e.currentTarget.checked)}
            />
          </InlineField>
        </InlineFieldRow>

        <InlineFieldRow>
          <InlineField
            label="With CA Cert"
            labelWidth={20}
            tooltip="Use a custom CA certificate for TLS verification"
          >
            <Switch
              value={jsonData.tlsWithCACert || false}
              onChange={(e) => onJsonDataChange('tlsWithCACert', e.currentTarget.checked)}
            />
          </InlineField>
        </InlineFieldRow>

        {jsonData.tlsWithCACert && (
          <InlineField label="CA Certificate" labelWidth={20} tooltip="PEM-encoded CA certificate">
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {secureJsonFields?.tlsCACert ? (
                <Input
                  width={40}
                  value="configured"
                  disabled
                  suffix={
                    <button type="button" className="btn btn-link" onClick={() => onResetSecureJsonData('tlsCACert')}>
                      Reset
                    </button>
                  }
                />
              ) : (
                <TextArea
                  cols={60}
                  rows={6}
                  value={secureJsonData?.tlsCACert || ''}
                  onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onSecureJsonDataChange('tlsCACert', e.target.value)}
                  placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                />
              )}
            </div>
          </InlineField>
        )}

        <InlineFieldRow>
          <InlineField
            label="With Client Cert"
            labelWidth={20}
            tooltip="Use client certificate for TLS authentication"
          >
            <Switch
              value={jsonData.tlsWithClientCert || false}
              onChange={(e) => onJsonDataChange('tlsWithClientCert', e.currentTarget.checked)}
            />
          </InlineField>
        </InlineFieldRow>

        {jsonData.tlsWithClientCert && (
          <>
            <InlineField label="Client Certificate" labelWidth={20} tooltip="PEM-encoded client certificate">
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {secureJsonFields?.tlsClientCert ? (
                  <Input
                    width={40}
                    value="configured"
                    disabled
                    suffix={
                      <button
                        type="button"
                        className="btn btn-link"
                        onClick={() => onResetSecureJsonData('tlsClientCert')}
                      >
                        Reset
                      </button>
                    }
                  />
                ) : (
                  <TextArea
                    cols={60}
                    rows={6}
                    value={secureJsonData?.tlsClientCert || ''}
                    onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
                      onSecureJsonDataChange('tlsClientCert', e.target.value)
                    }
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                  />
                )}
              </div>
            </InlineField>

            <InlineField label="Client Key" labelWidth={20} tooltip="PEM-encoded client private key">
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {secureJsonFields?.tlsClientKey ? (
                  <Input
                    width={40}
                    value="configured"
                    disabled
                    suffix={
                      <button
                        type="button"
                        className="btn btn-link"
                        onClick={() => onResetSecureJsonData('tlsClientKey')}
                      >
                        Reset
                      </button>
                    }
                  />
                ) : (
                  <TextArea
                    cols={60}
                    rows={6}
                    value={secureJsonData?.tlsClientKey || ''}
                    onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
                      onSecureJsonDataChange('tlsClientKey', e.target.value)
                    }
                    placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----"
                  />
                )}
              </div>
            </InlineField>
          </>
        )}
      </FieldSet>

      {/* Advanced Settings */}
      <FieldSet label="Advanced Settings">
        <InlineField
          label="Custom Query Parameters"
          labelWidth={20}
          tooltip="Additional query parameters to add to Prometheus requests (e.g., max_source_resolution=5m)"
        >
          <Input
            width={40}
            value={jsonData.customQueryParameters || ''}
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('customQueryParameters', e.target.value)}
            placeholder="param1=value1&param2=value2"
          />
        </InlineField>
      </FieldSet>
    </>
  );
}
