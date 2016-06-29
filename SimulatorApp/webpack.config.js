module.exports = {
  context: __dirname,
  entry: {
    javascript: "./components/app.js",
    html : "./index.html"
   },

  output: {
    filename: "app.js",
    path: __dirname + "/dist"
    
  },
  module : {
    loaders: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'babel',
        query : {
          presets : ['react']
        }
      }
    ]
  }
}

