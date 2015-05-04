var config = module.exports;

config["My tests"] = {
    rootPath: "..",
    environment: "browser", // or "node"
    sources: [
        "bower_components/jquery/dist/jquery.js",
    ],
    tests: [
        "diff/out/*.js"
    ],
    resources: [
        "diff/out/*.js.map"
    ]
}
