const {
  override,
  addWebpackAlias,
  addWebpackPlugin,
} = require("customize-cra");
const path = require("path");
const webpack = require("webpack");

module.exports = override(
  addWebpackAlias({
    crypto: path.resolve(__dirname, "node_modules/crypto-browserify"),
    stream: path.resolve(__dirname, "node_modules/stream-browserify"),
    assert: path.resolve(__dirname, "node_modules/assert"),
    process: "process/browser",
  }),
  addWebpackPlugin(
    new webpack.ProvidePlugin({
      process: "process/browser",
    })
  ),
  (config) => {
    config.resolve.fallback = {
      ...config.resolve.fallback,
      fs: false, // You can also use `require.resolve('browserify-fs')` if you need a polyfill
      path: require.resolve("path-browserify"),
      crypto: false,
    };
    return config;
  }
);

// module.exports = function override(config) {
//   config.resolve.fallback = {
//     ...config.resolve.fallback,
//     fs: false, // You can also use `require.resolve('browserify-fs')` if you need a polyfill
//     path: require.resolve("path-browserify"),
//     crypto: false,
//   };
//   return config;
// };
