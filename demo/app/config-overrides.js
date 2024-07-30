module.exports = function override(config) {
  config.resolve.fallback = {
    ...config.resolve.fallback,
    fs: false, // You can also use `require.resolve('browserify-fs')` if you need a polyfill
    path: require.resolve("path-browserify"),
  };
  return config;
};
