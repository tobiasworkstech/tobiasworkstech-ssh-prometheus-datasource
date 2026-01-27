/// <reference path="./plugins.d.ts" />
import CopyWebpackPlugin from 'copy-webpack-plugin';
import ForkTsCheckerWebpackPlugin from 'fork-ts-checker-webpack-plugin';
import { createRequire } from 'module';
import path from 'path';
import ReplaceInFileWebpackPlugin from 'replace-in-file-webpack-plugin';
import { fileURLToPath } from 'url';
import { Configuration } from 'webpack';

const require = createRequire(import.meta.url);
const __filename = fileURLToPath(import.meta.url);

const config = async (env: Record<string, unknown>): Promise<Configuration> => {
  const pluginJson = require('../../src/plugin.json');
  const pluginId = pluginJson.id;
  const isDevelopment = env.development === true;

  return {
    cache: {
      type: 'filesystem',
      buildDependencies: {
        config: [__filename],
      },
    },
    context: path.join(process.cwd(), 'src'),
    devtool: isDevelopment ? 'eval-source-map' : 'source-map',
    entry: './module.ts',
    externals: [
      'lodash',
      'react',
      'react-dom',
      '@grafana/data',
      '@grafana/runtime',
      '@grafana/ui',
      '@emotion/react',
      '@emotion/css',
    ],
    mode: isDevelopment ? 'development' : 'production',
    module: {
      rules: [
        {
          exclude: /node_modules/,
          test: /\.[tj]sx?$/,
          use: {
            loader: 'swc-loader',
            options: {
              jsc: {
                baseUrl: path.join(process.cwd(), 'src'),
                parser: {
                  syntax: 'typescript',
                  tsx: true,
                  decorators: true,
                  dynamicImport: true,
                },
                target: 'es2020',
              },
            },
          },
        },
        {
          test: /\.s?css$/,
          use: ['style-loader', 'css-loader', 'sass-loader'],
        },
        {
          test: /\.(png|jpe?g|gif|svg)$/,
          type: 'asset/resource',
          generator: {
            filename: 'images/[hash][ext]',
          },
        },
        {
          test: /\.(woff|woff2|eot|ttf|otf)(\?v=\d+\.\d+\.\d+)?$/,
          type: 'asset/resource',
          generator: {
            filename: 'fonts/[hash][ext]',
          },
        },
      ],
    },
    output: {
      clean: {
        keep: /gpx_/,
      },
      filename: 'module.js',
      library: {
        type: 'amd',
      },
      path: path.resolve(process.cwd(), 'dist'),
      publicPath: `public/plugins/${pluginId}/`,
    },
    plugins: [
      new CopyWebpackPlugin({
        patterns: [
          { from: 'plugin.json', to: '.' },
          { from: 'img/**/*', to: '.', noErrorOnMissing: true },
          { from: '../README.md', to: '.', noErrorOnMissing: true },
          { from: '../LICENSE', to: '.', noErrorOnMissing: true },
          { from: '../CHANGELOG.md', to: '.', noErrorOnMissing: true },
        ],
      }),
      new ForkTsCheckerWebpackPlugin({
        async: Boolean(isDevelopment),
        issue: {
          include: [{ file: '**/*.{ts,tsx}' }],
        },
        typescript: {
          configFile: path.join(process.cwd(), 'tsconfig.json'),
        },
      }),
      new ReplaceInFileWebpackPlugin([
        {
          dir: 'dist',
          files: ['plugin.json'],
          rules: [
            {
              search: /%VERSION%/g,
              replace: pluginJson.info?.version || '1.0.0',
            },
            {
              search: /%TODAY%/g,
              replace: new Date().toISOString().substring(0, 10),
            },
          ],
        },
      ]),
    ],
    resolve: {
      extensions: ['.ts', '.tsx', '.js', '.jsx'],
      unsafeCache: true,
    },
  };
};

export default config;
