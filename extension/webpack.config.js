const path = require("path");
const HappyPack = require("happypack");
const os = require("os");
const CopyPlugin = require("copy-webpack-plugin");

const srcDir = path.join(__dirname, "src");
const outDir = path.join(__dirname, "out");

const mainConfig = (plugins = []) => ({
	context: __dirname,
	devtool: "none",
	mode: "development",
	module: {
		rules: [
			{
				test: /\.scss$/,
				use: [
					"style-loader",
					"css-loader",
					"sass-loader",
				],
			},
			{
				use: [{
					loader: "happypack/loader?id=ts",
				}],
				test: /(^.?|\.[^d]|[^.]d|[^.][^d])\.tsx?$/,
			},
		],
	},
	plugins: [
		// new VueLoaderPlugin(),
		new HappyPack({
			id: "ts",
			threads: Math.max(os.cpus().length - 1, 1),
			loaders: [{
				path: "ts-loader",
				query: {
					happyPackMode: true,
				},
			}],
		}),
		...plugins,
	],
	resolve: {
		extensions: [".ts", ".tsx", ".js"]
	},
	stats: {
		all: false, // Fallback for options not defined.
		errors: true,
		warnings: true,
	},
});

module.exports = [
	{
		...mainConfig([
			new CopyPlugin(
				[
					{ from: path.join(srcDir, "config.html"), },
					{ from: path.join(__dirname, "logo128.png") },
					{ from: path.join(__dirname, "logo.svg") },
					{ from: path.join(__dirname, "manifest.json") },
					{ from: path.join(__dirname, "logo128.png") },
				],
				{
					copyUnmodified: true,
				}
			),
		]),
		entry: path.join(srcDir, "background.ts"),
		output: {
			path: outDir,
			filename: "background.js",
		},
	},
	{
		...mainConfig(),
		entry: path.join(srcDir, "content.ts"),
		output: {
			path: outDir,
			filename: "content.js",
		},
	},
	{
		...mainConfig(),
		entry: path.join(srcDir, "config.ts"),
		output: {
			path: outDir,
			filename: "config.js",
		},
	},
];
