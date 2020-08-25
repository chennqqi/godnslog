// webpack.config.js
module.exports = {
  module: {
    rules: [
      {
        test: /\.md$/i,
        use: 'raw-loader',
      },
    ],
  },
};