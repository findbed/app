import {
    interpolateExt,
    interpolate,
    extractJqPaths,
    buildDefaultObject,
    generateDefaultObject,
} from "./Interpolator";

describe("extended interpolate template", () => {
    it("with one empty object", () => {
        const template = "<p>${test} ${a.b}${a.d}</p>";
        const dataSource = {
            test: "hello",
            a: { b: "world", c: "!" },
        };
        const expected = "<p>hello world</p>";
        const actual = interpolateExt(template, dataSource);

        expect(actual).toEqual(expected);
    });

    it("with empty dataSource", () => {
        const template = "<p>${test} ${a.b}${a.d}</p>";
        const dataSource = {};
        const expected = "<p> </p>";
        const actual = interpolateExt(template, dataSource);

        expect(actual).toEqual(expected);
    });
});

describe("interpolate template", () => {
    it("with filled dataSource", () => {
        const template = "<p>${test} ${a.b}${a.d}</p>";
        const expected = ["test", "a.b", "a.d"];
        const actual = extractJqPaths(template);

        expect(actual).toEqual(expected);
    });
});

describe("extract jq-path from template", () => {
    it("with filled dataSource", () => {
        const jqpaths = ["test", "a.b", "a.c", "d.e"];
        const expected = {
            test: "",
            a: { b: "", c: "" },
            d: { e: "" },
        };
        const actual = buildDefaultObject(jqpaths);

        expect(actual).toEqual(expected);
    });
});

describe("build default data source from template", () => {
    it("with filled dataSource", () => {
        const template = "<p>${test}</p>";
        const dataSource = {
            test: "hello world",
        };
        const expected = "<p>hello world</p>";
        const actual = interpolate(template, dataSource);

        expect(actual).toEqual(expected);
    });
});

describe("generate default object", () => {
    it("simple object with primitive type", () => {
        const jqpath = "a";
        const defaultValue = "string";
        const expected = { a: defaultValue };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });

    it("very deep object with primitive type", () => {
        const jqpath = "a.b.c.d.e";
        const defaultValue = "string";
        const expected = {
            a: {
                b: {
                    c: {
                        d: {
                            e: defaultValue,
                        },
                    },
                },
            },
        };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });

    it("simple object with structure object", () => {
        const jqpath = "a";
        const defaultValue = { foo: "bar" };
        const expected = { a: defaultValue };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });

    it("very deep object with structure object", () => {
        const jqpath = "a.b.c.d.e";
        const defaultValue = { foo: "bar" };
        const expected = {
            a: {
                b: {
                    c: {
                        d: {
                            e: defaultValue,
                        },
                    },
                },
            },
        };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });

    it("simple object with structure array", () => {
        const jqpath = "a";
        const defaultValue = [1, 2, 3, 4, 5];
        const expected = { a: defaultValue };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });

    it("very deep object with structure array", () => {
        const jqpath = "a.b.c.d.e";
        const defaultValue = [1, 2, 3, 4, 5];
        const expected = {
            a: {
                b: {
                    c: {
                        d: {
                            e: defaultValue,
                        },
                    },
                },
            },
        };
        const actual = generateDefaultObject(jqpath, defaultValue);

        expect(actual).toEqual(expected);
    });
});
