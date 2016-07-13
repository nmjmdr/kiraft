var path = require('path');
var webpack = require('webpack');


module.exports = {
  context: __dirname,
  entry: "./main.js",
  
  output: {
    filename: 'bundle.js',
    path: __dirname    
  },
  module : {
    loaders: [
      {
        test: /.jsx?$/,
        exclude: /node_modules/,
        loader: 'babel-loader',
        query : {
          presets : ['es2015','react']
        }
      }
    ]
  }
}

