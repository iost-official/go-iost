const path = require('path');

module.exports = {
    mode: 'production',
    entry: './contract.js',
    output: {
        library: { type: 'commonjs2' },
        path: path.resolve(__dirname, 'dist'),
        filename: 'bundle.js',
    },
    target: 'node',
};
