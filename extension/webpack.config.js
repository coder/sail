const path = require("path");
const HappyPack = require("happypack");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const os = require("os");
const CopyPlugin = require("copy-webpack-plugin");
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
					//process.env.NODE_ENV !== "production" ? "style-loader" : MiniCssExtractPlugin.loader,
					MiniCssExtractPlugin.loader,
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
					{
						from: path.resolve(__dirname, "src/config.html"),
						to: path.resolve(process.cwd(), "out/config.html"),
					}
				],
				{
					copyUnmodified: true,
				}
			),
		]),
		entry: path.join(__dirname, "src", "background.ts"),
		output: {
			path: outDir,
			filename: "background.js",
		},
	},
	{
		...mainConfig(),
		entry: path.join(__dirname, "src", "content.ts"),
		output: {
			path: outDir,
			filename: "content.js",
		},
	},
	{
		...mainConfig(),
		entry: path.join(__dirname, "src", "config.ts"),
		output: {
			path: outDir,
			filename: "config.js",
		},
	},
	{
		...mainConfig([
			new MiniCssExtractPlugin({
				filename: "config.css",
				chunkFilename: "config.css"
			}),
		]),
		entry: path.join(__dirname, "src", "config.scss"),
		output: {
			path: outDir,
			filename: "config.css.js",
		},
	},
];
