import fs from "fs";
import path from "path";

import { getCardTemplate } from "./template";

test("getting a template", () => {
    const file = path.join(__dirname, "../web/views", "card.tmpl.html");
    const card = fs.readFileSync(file, "utf8", (_, data) => data);
    document.body.innerHTML = card;

    const actual = getCardTemplate();

    expect(actual).not.toEqual("");
});
