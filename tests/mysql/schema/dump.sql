
CREATE DATABASE findbed;

USE findbed;

CREATE TABLE timeslot_0001 (
    id bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    -- Required. CLDR region code of the country/region of the address. This
    -- is never inferred and it is up to the user to ensure the value is
    -- correct. See [Unicode CLDR Project](http://cldr.unicode.org/) and
    -- [Territory Information](http://www.unicode.org/cldr/charts/30/supplemental/territory_information.html)
    -- for details. Example: "CH" for Switzerland.
    region char(2) NOT NULL,
    -- Optional. Highest administrative subdivision which is used for postal
    -- addresses of a country or region.
    -- For example, this can be a state, a province, an oblast, or a prefecture.
    -- Specifically, for Spain this is the province and not the autonomous
    -- community (e.g. "Barcelona" and not "Catalonia").
    -- Many countries don't use an administrative area in postal addresses. E.g.
    -- in Switzerland this should be left unpopulated.
    area smallint(6) UNSIGNED NOT NULL,
    -- Optional. Generally refers to the city/town portion of the address.
    -- Examples: US city, IT comune, UK post town.
    -- In regions of the world where localities are not well defined or do not fit
    -- into this structure well, leave locality empty and use address_lines.
    locality smallint(6) UNSIGNED NOT NULL,
    -- Optional. Sublocality of the address.
    -- For example, this can be neighborhoods, boroughs, districts.
    sublocality smallint(6) UNSIGNED NOT NULL,
    housing_id bigint(20) UNSIGNED NOT NULL,
    lot_id bigint(20) UNSIGNED NOT NULL,
    start_at smallint(6) UNSIGNED DEFAULT 0,
    end_at smallint(6) UNSIGNED DEFAULT 65535,
    PRIMARY KEY (`id`),
    UNIQUE KEY slot (region, area, locality, sublocality, housing_id, lot_id, start_at),
    KEY free_slot (region, start_at, end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
