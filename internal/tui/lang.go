package tui

var translations = map[string]map[string]string{
	"en": {
		// Title
		"title": "TaxiCheck",

		// Main menu
		"main_menu":          "Main Menu",
		"main_calc":          " Calculate Fare",
		"main_settings":      " Settings",
		"main_help":          " Help/Manual",
		"main_setup":         " Initial Setup",
		"main_quit":          " Quit",
		"main_select":        "Select an option",

		// Setup
		"setup_title":         "Initial Setup",
		"setup_lang_title":    "Select Language",
		"setup_lang_en":       "English",
		"setup_lang_nl":       "Nederlands",
		"setup_lang_help":     "Use 1/2 to select, Enter to confirm",
		"setup_help":          "Tab: next field | Enter: save | Esc: cancel",
		"setup_board_fee":     " Board Fee",
		"setup_per_km":        " Per Km",
		"setup_per_minute":    " Per Minute",
		"setup_wait_minute":   " Wait Minute",

		// Settings
		"settings_title":      "Settings",
		"settings_help":       "Tab: next field | Enter: save | Esc: cancel",
		"settings_board_fee":  " Board Fee",
		"settings_per_km":     " Per Km",
		"settings_per_minute": " Per Minute",
		"settings_wait_minute": " Wait Minute",

		// Help
		"help_title":         "Help & Manual",
		"help_controls":      "Keyboard Controls:",
		"help_calc":          " - Calculate Fare",
		"help_settings":      " - Settings",
		"help_help":          " - Help/Manual",
		"help_setup":         " - Initial Setup",
		"help_quit":          " - Quit",
		"help_tab":           " - Next field",
		"help_enter":         " - Submit/Save",
		"help_esc":           " - Back",
		"help_config_title":  "Configuration is stored in:",
		"help_config_path":   "  ~/.taxiprijs/config.toml",
		"help_api_title":     "API (OpenStreetMap):",
		"help_api_desc":      "  Route calculation via OSRM & Nominatim",
		"help_pass_title":    "Passenger Groups:",
		"help_pass_1":        "  Taxi auto: max 4 passengers",
		"help_pass_2":        "  Taxi bus: 5-8 passengers",
		"help_pricing_title": "Pricing (per group):",
		"help_pricing_board": "  Board Fee: Initial charge",
		"help_pricing_km":    "  Per Km: Cost per kilometer",
		"help_pricing_time":  "  Per Minute: Cost per minute of ride",
		"help_return":        "Press Esc or Enter to return to main menu",

		// Calculate
		"calc_title":               "Calculate Fare",
		"calc_placeholder_start":   "e.g. Dam Square, Amsterdam",
		"calc_placeholder_end":     "e.g. Central Station, Rotterdam",
		"calc_placeholder_passengers": "1-8",
		"calc_label_start":         "Start address",
		"calc_label_end":           "Destination",
		"calc_label_passengers":    "Number of passengers (1-8)",
		"calc_mode":               "Route",
		"calc_mode_fastest":       "Fastest",
		"calc_mode_shortest":      "Shortest",
		"calc_help":               "Tab: next field | r: route mode | Enter: calculate | Esc: cancel",

		// Loading
		"loading": "Calculating route...",

		// Result
		"result_title":      "Fare Result",
		"result_route":      "Route:",
		"result_distance":   "Distance:  ",
		"result_duration":   "Duration:  ",
		"result_mode":       "Route mode: ",
		"result_group":      "Passenger Group: ",
		"result_board":      "Board Fee:  ",
		"result_km":         "Km Fee:     ",
		"result_time":       "Time Fee:   ",
		"result_total":      "Total:      ",
		"result_help":       "Press Enter or Esc to calculate again | q: quit",

		// Errors
		"err_invalid_input":    "Invalid input fields",
		"err_empty_start":      "Please enter a start address",
		"err_empty_end":        "Please enter a destination address",
		"err_invalid_pass":     "Passengers must be 1-8",
		"err_not_all_filled":   "Not all fields filled",
		"err_invalid_board":    "Invalid board fee for group ",
		"err_invalid_per_km":   "Invalid per km rate for group ",
		"err_invalid_per_min":  "Invalid per minute rate for group ",
		"err_invalid_wait_min": "Invalid wait minute rate for group ",
		"err_error":            "Error: ",
	},
	"nl": {
		// Title
		"title": "TaxiCheck",

		// Main menu
		"main_menu":          "Hoofdmenu",
		"main_calc":          " Prijs Berekenen",
		"main_settings":      " Instellingen",
		"main_help":          " Help/Handleiding",
		"main_setup":         " Eerste Installatie",
		"main_quit":          " Afsluiten",
		"main_select":        "Kies een optie",

		// Setup
		"setup_title":         "Installatie",
		"setup_lang_title":    "Kies Taal",
		"setup_lang_en":       "English",
		"setup_lang_nl":       "Nederlands",
		"setup_lang_help":     "Gebruik 1/2 om te kiezen, Enter om te bevestigen",
		"setup_help":          "Tab: volgend veld | Enter: opslaan | Esc: annuleren",
		"setup_board_fee":     " Instap",
		"setup_per_km":        " Per km",
		"setup_per_minute":    " Per min",
		"setup_wait_minute":   " Wacht",

		// Settings
		"settings_title":      "Instellingen",
		"settings_help":       "Tab: volgend veld | Enter: opslaan | Esc: annuleren",
		"settings_board_fee":  " Instap",
		"settings_per_km":     " Per km",
		"settings_per_minute": " Per min",
		"settings_wait_minute": " Wacht",

		// Help
		"help_title":         "Help & Handleiding",
		"help_controls":      "Toetsenbord Besturing:",
		"help_calc":          " - Prijs Berekenen",
		"help_settings":      " - Instellingen",
		"help_help":          " - Help/Handleiding",
		"help_setup":         " - Eerste Installatie",
		"help_quit":          " - Afsluiten",
		"help_tab":           " - Volgend veld",
		"help_enter":         " - Opslaan",
		"help_esc":           " - Terug",
		"help_config_title":  "Configuratie wordt opgeslagen in:",
		"help_config_path":   "  ~/.taxiprijs/config.toml",
		"help_api_title":     "API (OpenStreetMap):",
		"help_api_desc":      "  Routeberekening via OSRM & Nominatim",
		"help_pass_title":    "Passagiersgroepen:",
		"help_pass_1":        "  Taxi auto: max 4 passagiers",
		"help_pass_2":        "  Taxi bus: 5-8 passagiers",
		"help_pricing_title": "Tarieven (per groep):",
		"help_pricing_board": "  Instaptarief: Begin tarief",
		"help_pricing_km":    "  Per Km: Kosten per kilometer",
		"help_pricing_time":  "  Per Minuut: Kosten per minuut rit",
		"help_return":        "Druk op Esc of Enter om terug te gaan naar het hoofdmenu",

		// Calculate
		"calc_title":               "Prijs Berekenen",
		"calc_placeholder_start":   "bijv. Dam, Amsterdam",
		"calc_placeholder_end":     "bijv. Centraal Station, Rotterdam",
		"calc_placeholder_passengers": "1-8",
		"calc_label_start":         "Vertrekadres",
		"calc_label_end":           "Bestemming",
		"calc_label_passengers":    "Aantal passagiers (1-8)",
		"calc_mode":               "Route",
		"calc_mode_fastest":       "Snelste",
		"calc_mode_shortest":      "Kortste",
		"calc_help":               "Tab: volgend veld | r: route type | Enter: berekenen | Esc: annuleren",

		// Loading
		"loading": "Route berekenen...",

		// Result
		"result_title":      "Prijs Resultaat",
		"result_route":      "Route:",
		"result_distance":   "Afstand:   ",
		"result_duration":   "Duur:      ",
		"result_mode":       "Route type: ",
		"result_group":      "Passagiersgroep: ",
		"result_board":      "Instaptarief:  ",
		"result_km":         "Km kosten:     ",
		"result_time":       "Tijdkosten:    ",
		"result_total":      "Totaal:        ",
		"result_help":       "Druk op Enter of Esc om opnieuw te berekenen | q: afsluiten",

		// Errors
		"err_invalid_input":    "Ongeldige invoervelden",
		"err_empty_start":      "Voer een vertrekadres in",
		"err_empty_end":        "Voer een bestemming in",
		"err_invalid_pass":     "Passagiers moeten 1-8 zijn",
		"err_not_all_filled":   "Niet alle velden zijn ingevuld",
		"err_invalid_board":    "Ongeldig instaptarief voor groep ",
		"err_invalid_per_km":   "Ongeldig per km tarief voor groep ",
		"err_invalid_per_min":  "Ongeldig per minuut tarief voor groep ",
		"err_invalid_wait_min": "Ongeldig wachtminuut tarief voor groep ",
		"err_error":            "Fout: ",
	},
}

func t(lang, key string) string {
	if lang != "nl" {
		lang = "en"
	}
	if val, ok := translations[lang][key]; ok {
		return val
	}
	if val, ok := translations["en"][key]; ok {
		return val
	}
	return key
}
