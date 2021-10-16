module.exports = {
    publicPath: '/ui/',
    outputDir: '../internal/web/dist',
    css: {
        requireModuleExtension: false,
    },
    pages: {
        console: {
            entry: "src/console.main.js",
            template: "public/index.html",
            filename: "index.html",
            title: "AetherFS",
            chunks: ["vendor", "console-vendor", "console"],
        },
        swagger: {
            entry: "src/swagger.main.js",
            template: "public/index.html",
            filename: "swagger/index.html",
            title: "AetherFS Swagger",
            chunks: ["vendor", "swagger-vendor", "swagger"],
        },
    },
    chainWebpack: config => {
        const options = module.exports
        const pages = options.pages
        const pageKeys = Object.keys(pages)

        // Long-term caching

        const IS_VENDOR = /[\\/]node_modules[\\/]/

        config.optimization
            .splitChunks({
                cacheGroups: {
                    ...pageKeys.map(key => ({
                        name: `${key}-vendor`,
                        priority: -11,
                        chunks: chunk => chunk.name === key,
                        test: IS_VENDOR,
                        enforce: true,
                    })),
                    vendor: {
                        name: 'vendor',
                        priority: -20,
                        chunks: 'initial',
                        minChunks: 2,
                        reuseExistingChunk: true,
                        enforce: true,
                    },
                },
            })
    },
}
