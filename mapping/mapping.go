package mapping

var NumberWords = map[string]int{
	"zero":  0,
	"one":   1,
	"two":   2,
	"three": 3,
	"four":  4,
	"five":  5,
	"six":   6,
	"seven": 7,
	"eight": 8,
	"nine":  9,
	"ten":   10,
}

var Country3ToLanguage = map[string]string{
	"AFG": "pus", // Afghanistan -> Pashto
	"ALB": "sqi", // Albania -> Albanian
	"DZA": "ara", // Algeria -> Arabic
	"ASM": "eng", // American Samoa -> English
	"AND": "cat", // Andorra -> Catalan
	"AGO": "por", // Angola -> Portuguese
	"AIA": "eng", // Anguilla -> English
	"ATA": "eng", // Antarctica -> English
	"ATG": "eng", // Antigua and Barbuda -> English
	"ARG": "spa", // Argentina -> Spanish
	"ARM": "hye", // Armenia -> Armenian
	"ABW": "nld", // Aruba -> Dutch
	"AUS": "eng", // Australia -> English
	"AUT": "deu", // Austria -> German
	"AZE": "aze", // Azerbaijan -> Azerbaijani
	"BHS": "eng", // Bahamas -> English
	"BHR": "ara", // Bahrain -> Arabic
	"BGD": "ben", // Bangladesh -> Bengali
	"BRB": "eng", // Barbados -> English
	"BLR": "bel", // Belarus -> Belarusian
	"BEL": "nld", // Belgium -> Dutch
	"BLZ": "eng", // Belize -> English
	"BEN": "fra", // Benin -> French
	"BMU": "eng", // Bermuda -> English
	"BTN": "dzo", // Bhutan -> Dzongkha
	"BOL": "spa", // Bolivia -> Spanish
	"BIH": "bos", // Bosnia and Herzegovina -> Bosnian
	"BWA": "eng", // Botswana -> English
	"BRA": "por", // Brazil -> Portuguese
	"BRN": "msa", // Brunei -> Malay
	"BGR": "bul", // Bulgaria -> Bulgarian
	"BFA": "fra", // Burkina Faso -> French
	"BDI": "fra", // Burundi -> French
	"KHM": "khm", // Cambodia -> Khmer
	"CMR": "fra", // Cameroon -> French
	"CAN": "eng", // Canada -> English
	"CPV": "por", // Cape Verde -> Portuguese
	"CYM": "eng", // Cayman Islands -> English
	"CAF": "fra", // Central African Republic -> French
	"TCD": "ara", // Chad -> Arabic
	"CHL": "spa", // Chile -> Spanish
	"CHN": "zho", // China -> Chinese
	"COL": "spa", // Colombia -> Spanish
	"COM": "ara", // Comoros -> Arabic
	"COG": "fra", // Congo -> French
	"COD": "fra", // Democratic Republic of the Congo -> French
	"CRI": "spa", // Costa Rica -> Spanish
	"CIV": "fra", // Côte d'Ivoire -> French
	"HRV": "hrv", // Croatia -> Croatian
	"CUB": "spa", // Cuba -> Spanish
	"CYP": "ell", // Cyprus -> Greek
	"CZE": "ces", // Czech Republic -> Czech
	"DNK": "dan", // Denmark -> Danish
	"DJI": "ara", // Djibouti -> Arabic
	"DMA": "eng", // Dominica -> English
	"DOM": "spa", // Dominican Republic -> Spanish
	"ECU": "spa", // Ecuador -> Spanish
	"EGY": "ara", // Egypt -> Arabic
	"SLV": "spa", // El Salvador -> Spanish
	"GNQ": "spa", // Equatorial Guinea -> Spanish
	"ERI": "tir", // Eritrea -> Tigrinya
	"EST": "est", // Estonia -> Estonian
	"SWZ": "eng", // Eswatini -> English
	"ETH": "amh", // Ethiopia -> Amharic
	"FJI": "eng", // Fiji -> English
	"FIN": "fin", // Finland -> Finnish
	"FRA": "fra", // France -> French
	"GAB": "fra", // Gabon -> French
	"GMB": "eng", // Gambia -> English
	"GEO": "kat", // Georgia -> Georgian
	"DEU": "deu", // Germany -> German
	"GHA": "eng", // Ghana -> English
	"GRC": "ell", // Greece -> Greek
	"GRD": "eng", // Grenada -> English
	"GUM": "eng", // Guam -> English
	"GTM": "spa", // Guatemala -> Spanish
	"GIN": "fra", // Guinea -> French
	"GNB": "por", // Guinea-Bissau -> Portuguese
	"GUY": "eng", // Guyana -> English
	"HTI": "hat", // Haiti -> Haitian Creole
	"HND": "spa", // Honduras -> Spanish
	"HUN": "hun", // Hungary -> Hungarian
	"ISL": "isl", // Iceland -> Icelandic
	"IND": "hin", // India -> Hindi
	"IDN": "ind", // Indonesia -> Indonesian
	"IRN": "fas", // Iran -> Persian
	"IRQ": "ara", // Iraq -> Arabic
	"IRL": "gle", // Ireland -> Irish
	"ISR": "heb", // Israel -> Hebrew
	"ITA": "ita", // Italy -> Italian
	"JAM": "eng", // Jamaica -> English
	"JPN": "jpn", // Japan -> Japanese
	"JOR": "ara", // Jordan -> Arabic
	"KAZ": "kaz", // Kazakhstan -> Kazakh
	"KEN": "eng", // Kenya -> English
	"KIR": "eng", // Kiribati -> English
	"PRK": "kor", // North Korea -> Korean
	"KOR": "kor", // South Korea -> Korean
	"KWT": "ara", // Kuwait -> Arabic
	"KGZ": "kir", // Kyrgyzstan -> Kyrgyz
	"LAO": "lao", // Laos -> Lao
	"LVA": "lav", // Latvia -> Latvian
	"LBN": "ara", // Lebanon -> Arabic
	"LSO": "eng", // Lesotho -> English
	"LBR": "eng", // Liberia -> English
	"LBY": "ara", // Libya -> Arabic
	"LIE": "deu", // Liechtenstein -> German
	"LTU": "lit", // Lithuania -> Lithuanian
	"LUX": "ltz", // Luxembourg -> Luxembourgish
	"MDG": "mlg", // Madagascar -> Malagasy
	"MWI": "eng", // Malawi -> English
	"MYS": "msa", // Malaysia -> Malay
	"MDV": "div", // Maldives -> Dhivehi
	"MLI": "fra", // Mali -> French
	"MLT": "mlt", // Malta -> Maltese
	"MHL": "eng", // Marshall Islands -> English
	"MRT": "ara", // Mauritania -> Arabic
	"MUS": "eng", // Mauritius -> English
	"MEX": "spa", // Mexico -> Spanish
	"FSM": "eng", // Micronesia -> English
	"MDA": "ron", // Moldova -> Romanian
	"MCO": "fra", // Monaco -> French
	"MNG": "mon", // Mongolia -> Mongolian
	"MNE": "srp", // Montenegro -> Serbian
	"MAR": "ara", // Morocco -> Arabic
	"MOZ": "por", // Mozambique -> Portuguese
	"MMR": "mya", // Myanmar -> Burmese
	"NAM": "eng", // Namibia -> English
	"NRU": "eng", // Nauru -> English
	"NPL": "nep", // Nepal -> Nepali
	"NLD": "nld", // Netherlands -> Dutch
	"NZL": "eng", // New Zealand -> English
	"NIC": "spa", // Nicaragua -> Spanish
	"NER": "fra", // Niger -> French
	"NGA": "eng", // Nigeria -> English
	"MKD": "mkd", // North Macedonia -> Macedonian
	"NOR": "nob", // Norway -> Norwegian
	"OMN": "ara", // Oman -> Arabic
	"PAK": "urd", // Pakistan -> Urdu
	"PLW": "eng", // Palau -> English
	"PAN": "spa", // Panama -> Spanish
	"PNG": "eng", // Papua New Guinea -> English
	"PRY": "spa", // Paraguay -> Spanish
	"PER": "spa", // Peru -> Spanish
	"PHL": "fil", // Philippines -> Filipino
	"POL": "pol", // Poland -> Polish
	"PRT": "por", // Portugal -> Portuguese
	"QAT": "ara", // Qatar -> Arabic
	"ROU": "ron", // Romania -> Romanian
	"RUS": "rus", // Russia -> Russian
	"RWA": "kin", // Rwanda -> Kinyarwanda
	"WSM": "eng", // Samoa -> English
	"SMR": "ita", // San Marino -> Italian
	"STP": "por", // Sao Tome and Principe -> Portuguese
	"SAU": "ara", // Saudi Arabia -> Arabic
	"SEN": "fra", // Senegal -> French
	"SRB": "srp", // Serbia -> Serbian
	"SYC": "eng", // Seychelles -> English
	"SLE": "eng", // Sierra Leone -> English
	"SGP": "zho", // Singapore -> Chinese
	"SVK": "slk", // Slovakia -> Slovak
	"SVN": "slv", // Slovenia -> Slovenian
	"SLB": "eng", // Solomon Islands -> English
	"SOM": "som", // Somalia -> Somali
	"ZAF": "eng", // South Africa -> English
	"SSD": "eng", // South Sudan -> English
	"ESP": "spa", // Spain -> Spanish
	"LKA": "sin", // Sri Lanka -> Sinhala
	"SDN": "ara", // Sudan -> Arabic
	"SUR": "nld", // Suriname -> Dutch
	"SWE": "swe", // Sweden -> Swedish
	"CHE": "deu", // Switzerland -> German
	"SYR": "ara", // Syria -> Arabic
	"TWN": "zho", // Taiwan -> Chinese
	"TJK": "tgk", // Tajikistan -> Tajik
	"TZA": "swa", // Tanzania -> Swahili
	"THA": "tha", // Thailand -> Thai
	"TLS": "tet", // Timor-Leste -> Tetum
	"TGO": "fra", // Togo -> French
	"TON": "eng", // Tonga -> English
	"TTO": "eng", // Trinidad and Tobago -> English
	"TUN": "ara", // Tunisia -> Arabic
	"TUR": "tur", // Turkey -> Turkish
	"TKM": "tuk", // Turkmenistan -> Turkmen
	"TUV": "eng", // Tuvalu -> English
	"UGA": "eng", // Uganda -> English
	"UKR": "ukr", // Ukraine -> Ukrainian
	"ARE": "ara", // United Arab Emirates -> Arabic
	"GBR": "eng", // United Kingdom -> English
	"USA": "eng", // United States -> English
	"URY": "spa", // Uruguay -> Spanish
	"UZB": "uzb", // Uzbekistan -> Uzbek
	"VUT": "bis", // Vanuatu -> Bislama
	"VAT": "ita", // Vatican City -> Italian
	"VEN": "spa", // Venezuela -> Spanish
	"VNM": "vie", // Vietnam -> Vietnamese
	"YEM": "ara", // Yemen -> Arabic
	"ZMB": "eng", // Zambia -> English
	"ZWE": "eng", // Zimbabwe -> English
}

var Iso3ToIso2 = map[string]string{
	"AFG": "AF",
	"ALB": "AL",
	"DZA": "DZ",
	"ASM": "AS",
	"AND": "AD",
	"AGO": "AO",
	"AIA": "AI",
	"ATA": "AQ",
	"ATG": "AG",
	"ARG": "AR",
	"ARM": "AM",
	"ABW": "AW",
	"AUS": "AU",
	"AUT": "AT",
	"AZE": "AZ",
	"BHS": "BS",
	"BHR": "BH",
	"BGD": "BD",
	"BRB": "BB",
	"BLR": "BY",
	"BEL": "BE",
	"BLZ": "BZ",
	"BEN": "BJ",
	"BMU": "BM",
	"BTN": "BT",
	"BOL": "BO",
	"BES": "BQ",
	"BIH": "BA",
	"BWA": "BW",
	"BVT": "BV",
	"BRA": "BR",
	"IOT": "IO",
	"BRN": "BN",
	"BGR": "BG",
	"BFA": "BF",
	"BDI": "BI",
	"CPV": "CV",
	"KHM": "KH",
	"CMR": "CM",
	"CAN": "CA",
	"CYM": "KY",
	"CAF": "CF",
	"TCD": "TD",
	"CHL": "CL",
	"CHN": "CN",
	"CXR": "CX",
	"CCK": "CC",
	"COL": "CO",
	"COM": "KM",
	"COG": "CG",
	"COD": "CD",
	"COK": "CK",
	"CRI": "CR",
	"HRV": "HR",
	"CUB": "CU",
	"CUW": "CW",
	"CYP": "CY",
	"CZE": "CZ",
	"DNK": "DK",
	"DJI": "DJ",
	"DMA": "DM",
	"DOM": "DO",
	"ECU": "EC",
	"EGY": "EG",
	"SLV": "SV",
	"GNQ": "GQ",
	"ERI": "ER",
	"EST": "EE",
	"SWZ": "SZ",
	"ETH": "ET",
	"FLK": "FK",
	"FRO": "FO",
	"FJI": "FJ",
	"FIN": "FI",
	"FRA": "FR",
	"GUF": "GF",
	"PYF": "PF",
	"ATF": "TF",
	"GAB": "GA",
	"GMB": "GM",
	"GEO": "GE",
	"DEU": "DE",
	"GHA": "GH",
	"GIB": "GI",
	"GRC": "GR",
	"GRL": "GL",
	"GRD": "GD",
	"GLP": "GP",
	"GUM": "GU",
	"GTM": "GT",
	"GGY": "GG",
	"GIN": "GN",
	"GNB": "GW",
	"GUY": "GY",
	"HTI": "HT",
	"HMD": "HM",
	"VAT": "VA",
	"HND": "HN",
	"HKG": "HK",
	"HUN": "HU",
	"ISL": "IS",
	"IND": "IN",
	"IDN": "ID",
	"IRN": "IR",
	"IRQ": "IQ",
	"IRL": "IE",
	"IMN": "IM",
	"ISR": "IL",
	"ITA": "IT",
	"JAM": "JM",
	"JPN": "JP",
	"JEY": "JE",
	"JOR": "JO",
	"KAZ": "KZ",
	"KEN": "KE",
	"KIR": "KI",
	"PRK": "KP",
	"KOR": "KR",
	"KWT": "KW",
	"KGZ": "KG",
	"LAO": "LA",
	"LVA": "LV",
	"LBN": "LB",
	"LSO": "LS",
	"LBR": "LR",
	"LBY": "LY",
	"LIE": "LI",
	"LTU": "LT",
	"LUX": "LU",
	"MAC": "MO",
	"MDG": "MG",
	"MWI": "MW",
	"MYS": "MY",
	"MDV": "MV",
	"MLI": "ML",
	"MLT": "MT",
	"MHL": "MH",
	"MTQ": "MQ",
	"MRT": "MR",
	"MUS": "MU",
	"MYT": "YT",
	"MEX": "MX",
	"FSM": "FM",
	"MDA": "MD",
	"MCO": "MC",
	"MNG": "MN",
	"MNE": "ME",
	"MSR": "MS",
	"MAR": "MA",
	"MOZ": "MZ",
	"MMR": "MM",
	"NAM": "NA",
	"NRU": "NR",
	"NPL": "NP",
	"NLD": "NL",
	"NCL": "NC",
	"NZL": "NZ",
	"NIC": "NI",
	"NER": "NE",
	"NGA": "NG",
	"NIU": "NU",
	"NFK": "NF",
	"MKD": "MK",
	"MNP": "MP",
	"NOR": "NO",
	"OMN": "OM",
	"PAK": "PK",
	"PLW": "PW",
	"PSE": "PS",
	"PAN": "PA",
	"PNG": "PG",
	"PRY": "PY",
	"PER": "PE",
	"PHL": "PH",
	"PCN": "PN",
	"POL": "PL",
	"PRT": "PT",
	"PRI": "PR",
	"QAT": "QA",
	"REU": "RE",
	"ROU": "RO",
	"RUS": "RU",
	"RWA": "RW",
	"BLM": "BL",
	"SHN": "SH",
	"KNA": "KN",
	"LCA": "LC",
	"MAF": "MF",
	"SPM": "PM",
	"VCT": "VC",
	"WSM": "WS",
	"SMR": "SM",
	"STP": "ST",
	"SAU": "SA",
	"SEN": "SN",
	"SRB": "RS",
	"SYC": "SC",
	"SLE": "SL",
	"SGP": "SG",
	"SXM": "SX",
	"SVK": "SK",
	"SVN": "SI",
	"SLB": "SB",
	"SOM": "SO",
	"ZAF": "ZA",
	"SGS": "GS",
	"SSD": "SS",
	"ESP": "ES",
	"LKA": "LK",
	"SDN": "SD",
	"SUR": "SR",
	"SJM": "SJ",
	"SWE": "SE",
	"CHE": "CH",
	"SYR": "SY",
	"TWN": "TW",
	"TJK": "TJ",
	"TZA": "TZ",
	"THA": "TH",
	"TLS": "TL",
	"TGO": "TG",
	"TKL": "TK",
	"TON": "TO",
	"TTO": "TT",
	"TUN": "TN",
	"TUR": "TR",
	"TKM": "TM",
	"TCA": "TC",
	"TUV": "TV",
	"UGA": "UG",
	"UKR": "UA",
	"ARE": "AE",
	"GBR": "GB",
	"UMI": "UM",
	"USA": "US",
	"URY": "UY",
	"UZB": "UZ",
	"VUT": "VU",
	"VEN": "VE",
	"VNM": "VN",
	"VGB": "VG",
	"VIR": "VI",
	"WLF": "WF",
	"ESH": "EH",
	"YEM": "YE",
	"ZMB": "ZM",
	"ZWE": "ZW",
}

var PlaceRank = map[string]int{
	"city":              900,
	"borough":           850,
	"suburb":            450,
	"quarter":           845,
	"neighbourhood":     815,
	"city_block":        800,
	"plot":              500,
	"town":              850,
	"village":           800,
	"hamlet":            750,
	"isolated_dwelling": 300,
	"farm":              300,
	"allotments":        250,
}

var Language3ToLanguage2 = map[string]string{
	"aar": "aa",
	"abk": "ab",
	"afr": "af",
	"aka": "ak",
	"amh": "am",
	"ara": "ar",
	"arg": "an",
	"asm": "as",
	"ava": "av",
	"ave": "ae",
	"aym": "ay",
	"aze": "az",
	"bak": "ba",
	"bam": "bm",
	"bel": "be",
	"ben": "bn",
	"bis": "bi",
	"bod": "bo",
	"bos": "bs",
	"bre": "br",
	"bul": "bg",
	"cat": "ca",
	"ces": "cs",
	"cha": "ch",
	"che": "ce",
	"chu": "cu",
	"chv": "cv",
	"cor": "kw",
	"cos": "co",
	"cre": "cr",
	"cym": "cy",
	"dan": "da",
	"deu": "de",
	"div": "dv",
	"dzo": "dz",
	"ell": "el",
	"eng": "en",
	"epo": "eo",
	"est": "et",
	"ewe": "ee",
	"fao": "fo",
	"fas": "fa",
	"fij": "fj",
	"fin": "fi",
	"fra": "fr",
	"fry": "fy",
	"ful": "ff",
	"kat": "ka",
	"glg": "gl",
	"lug": "lg",
	"glv": "gv",
	"slk": "sk",
	"grn": "gn",
	"guj": "gu",
	"hat": "ht",
	"hau": "ha",
	"heb": "he",
	"her": "hz",
	"hin": "hi",
	"hun": "hu",
	"hye": "hy",
	"iba": "iba",
	"ibo": "ig",
	"ind": "id",
	"ina": "ia",
	"isl": "is",
	"ita": "it",
	"iku": "iu",
	"jpn": "ja",
	"jav": "jv",
	"kal": "kl",
	"kan": "kn",
	"kas": "ks",
	"kau": "kr",
	"kaz": "kk",
	"khm": "km",
	"kik": "ki",
	"kin": "rw",
	"kir": "ky",
	"kom": "kv",
	"kon": "kg",
	"kor": "ko",
	"kua": "kj",
	"kur": "ku",
	"lao": "lo",
	"lat": "la",
	"lav": "lv",
	"lim": "li",
	"lin": "ln",
	"lit": "lt",
	"ltz": "lb",
	"lub": "lu",
	"mkd": "mk",
	"mlg": "mg",
	"msa": "ms",
	"mal": "ml",
	"mlt": "mt",
	"mri": "mi",
	"mar": "mr",
	"mhk": "mh",
	"mon": "mn",
	"nau": "na",
	"nav": "nv",
	"nbl": "nr",
	"nde": "nd",
	"ndo": "ng",
	"nep": "ne",
	"nno": "nn",
	"nob": "nb",
	"nor": "no",
	"nya": "ny",
	"oci": "oc",
	"oji": "oj",
	"ori": "or",
	"orm": "om",
	"oss": "os",
	"pan": "pa",
	"pli": "pi",
	"pol": "pl",
	"pus": "ps",
	"por": "pt",
	"que": "qu",
	"roh": "rm",
	"ron": "ro",
	"run": "rn",
	"rus": "ru",
	"sag": "sg",
	"san": "sa",
	"sin": "si",
	"slv": "sl",
	"som": "so",
	"sot": "st",
	"spa": "es",
	"srd": "sc",
	"srp": "sr",
	"ssw": "ss",
	"swa": "sw",
	"swe": "sv",
	"tah": "ty",
	"tam": "ta",
	"tat": "tt",
	"tel": "te",
	"tgk": "tg",
	"tgl": "tl",
	"tha": "th",
	"tir": "ti",
	"ton": "to",
	"tsn": "tn",
	"tso": "ts",
	"tuk": "tk",
	"tur": "tr",
	"twi": "tw",
	"uig": "ug",
	"ukr": "uk",
	"urd": "ur",
	"uzb": "uz",
	"ven": "ve",
	"vie": "vi",
	"vol": "vo",
	"wln": "wa",
	"wol": "wo",
	"xho": "xh",
	"yid": "yi",
	"yor": "yo",
	"zha": "za",
	"zho": "zh",
	"zul": "zu",
}
