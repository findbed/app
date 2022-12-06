// Copyright © 2022 Dmitry Stoletov <info@imega.ru>
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

import fetch from "./fetch";
import { getCardTemplate } from "./template";
import { interpolateExt } from "./Interpolator";

const cardTemplate = getCardTemplate();
const main = document.getElementById("main");

fetch().then((data) => {
    data.data.forEach((item) => {
        const card = document.createElement("div");
        card.innerHTML = interpolateExt(cardTemplate, item);
        main.appendChild(card);
    });
});
