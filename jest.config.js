"use strict";

const config = {
    testPathIgnorePatterns: [".*.skip.test.js", "/node_modules/"],
    verbose: true,
    testEnvironment: "jsdom",
};

module.exports = config;
