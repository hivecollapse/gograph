package template

import (
	"gograph/internal/log"
	"os"
	"strings"
	gotemplate "text/template"

	rdata "github.com/Pallinder/go-randomdata"
)

// Read an environment variable
//
//	Usage:
//	- { "token": "{{ env "ENV_TOKEN" }}" }
//	- Bearer {{ "ENV_TOKEN" | env  }}
func env(envvar string, def ...string) string {
	value, exists := os.LookupEnv(envvar)
	if !exists && len(def) > 0 {
		return def[0]
	}
	return value
}

// Map of extra function for the template
var funcMap = gotemplate.FuncMap{
	"env": env,

	"randSillyName": func() string { return rdata.SillyName() },

	// Print a male title
	"randomTitleMale": func() string { return rdata.Title(rdata.Male) },

	// Print a female title
	"randomTitleFemale": func() string { return rdata.Title(rdata.Female) },

	// Print a title with random gender
	"randomTitle": func() string { return rdata.Title(rdata.RandomGender) },

	// Print a male first name
	"randomFirstNameMale": func() string { return rdata.FirstName(rdata.Male) },

	// Print a female first name
	"randomFirstNameFemale": func() string { return rdata.FirstName(rdata.Female) },

	// Print a firstname with random gender
	"randomFirstName": func() string { return rdata.FirstName(rdata.RandomGender) },

	// Print a last name
	"randomLastName": func() string { return rdata.LastName() },

	// Print a male name
	"randomFullNameMale": func() string { return rdata.FullName(rdata.Male) },

	// Print a female name
	"randomFullNameFemale": func() string { return rdata.FullName(rdata.Female) },

	// Print a name with random gender
	"randomFullName": func() string { return rdata.FullName(rdata.RandomGender) },

	// Print an email
	"randomEmail": func() string { return rdata.Email() },

	// Print a country with full text representation
	"randomCountry": func() string { return rdata.Country(rdata.FullCountry) },

	// Print a country using ISO 3166-1 alpha-2
	"randomCountryAlpha2": func() string { return rdata.Country(rdata.TwoCharCountry) },

	// Print a country using ISO 3166-1 alpha-3
	"randomCountryAlpha3": func() string { return rdata.Country(rdata.ThreeCharCountry) },

	// Print BCP 47 language tag
	"randomLocale": func() string { return rdata.Locale() },

	// Print a currency using ISO 4217
	"randomCurrency": func() string { return rdata.Currency() },

	// Print the name of a random city
	"randomCity": func() string { return rdata.City() },

	// Print the name of a random american state
	"randomUSState": func() string { return rdata.State(rdata.Large) },

	// Print the name of a random american state using two chars
	"randomUSState2": func() string { return rdata.State(rdata.Small) },

	// Print an american sounding street name
	"randomStreet": func() string { return rdata.Street() },

	// Print an american sounding address
	"randomAddress": func() string { return rdata.Address() },

	// Print a random number between 0 and 100 included
	"randomNumber100": func() int { return rdata.Number(0, 101) },

	// Print a random number >= x and < y
	"randomNumber": func(a ...int) int { return rdata.Number(a...) },

	// Print a random float with 2 decimanl between 0 and 100 included
	"randomDecimal100": func() float64 { return rdata.Decimal(0, 101, 2) },

	// Print a random float >= x and < y and z decimal
	"randomDecimal": func(a ...int) float64 { return rdata.Decimal(a...) },

	// Print a bool
	"randomBoolean": func() bool { return rdata.Boolean() },

	// Print a paragraph
	"randomParagraph": func() string { return rdata.Paragraph() },

	// Print a postal code
	"randomPostalCode": func(x string) string { return rdata.PostalCode(x) },

	// Return a random string of len n
	"randomString": func(len int) string { return rdata.RandStringRunes(len) },

	// Print a set of 2 random numbers as a string
	"randomStringNumber": func(numberPairs int, separator string) string { return rdata.StringNumber(numberPairs, separator) },

	// Print a random string sampled from a list of strings
	"randomStringSample": func(a ...string) string { return rdata.StringSample(a...) },

	// Print a valid random IPv4 address
	"randomIpV4": func() string { return rdata.IpV4Address() },

	// Print a valid random IPv6 address
	"randomIpV6": func() string { return rdata.IpV6Address() },

	// Print a browser's user agent string
	"randomUserAgentString": func() string { return rdata.UserAgentString() },

	// Print a day
	"randomDay": func() string { return rdata.Day() },

	// Print a month
	"randomMonth": func() string { return rdata.Month() },

	// Print full date like Monday 22 Aug 2016
	"randomFullDate": func() string { return rdata.FullDate() },

	// Print full date <= Monday 22 Aug 2016
	"randomFullDateInRange": func(a ...string) string { return rdata.FullDateInRange(a...) },

	// Print phone number according to e.164
	"randomPhoneNumber": func() string { return rdata.PhoneNumber() },

	// Get a random country-localised street name for Great Britain => GB
	"randomStreetForCountry": func(country string) string { return rdata.StreetForCountry(country) },

	// Get a random country-localised province for Great Britain => GB
	"randomProvinceForCountry": func(country string) string { return rdata.ProvinceForCountry(country) },
}

// Run a template returning errors if any
func RunTemplate(text string, context any) (string, error) {

	tmpl := gotemplate.New("")

	// Setup custom function

	tmpl.Funcs(funcMap)

	tmpl, err := tmpl.Parse(text)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	err = tmpl.Execute(&b, context)
	if err != nil {
		return "", err
	}
	r := strings.TrimSpace(b.String())

	log.Debugf("Template parsed:\n----input----\n%v\n----output----\n%v\n--------\n", text, r)

	return r, nil
}

// Run a template returning the input value on error
func RunTemplateOrUnparsed(text string, context any) string {
	out, err := RunTemplate(text, context)
	if err != nil {
		log.Println("Failed to parse template", err)
		return text
	}
	return out
}
