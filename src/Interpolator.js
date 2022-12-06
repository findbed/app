// Copyright © 2020 Dmitry Stoletov <info@imega.ru>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import merge from "./lodash/merge";

const interpolate = (template, data_source = {}) => {
    const names = Object.keys(data_source);
    const vals = Object.values(data_source);

    return new Function(...names, `return \`${template}\`;`)(...vals);
};

const interpolateExt = (template, data_source = {}) => {
    const ds = buildDefaultObject(extractJqPaths(template));
    const union = merge(ds, data_source);

    return interpolate(template, union);
};

const extractJqPaths = (template) => {
    const matches = [];

    for (const m of template.matchAll(/\${([^}]+)}/g)) {
        matches.push(m[1]);
    }

    return matches;
};

const buildDefaultObject = (jqpaths) => {
    let res = {};

    jqpaths.forEach((i) => {
        res = merge(res, generateDefaultObject(i, ""));
    });

    return res;
};

const generateDefaultObject = (jqpath, defaultValue) =>
    jqpath.split(".").reduceRight((a, c) => ({ [c]: a }), defaultValue);

export {
    interpolateExt,
    interpolate,
    extractJqPaths,
    buildDefaultObject,
    generateDefaultObject,
};
