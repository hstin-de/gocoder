package generate

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"hstin/gocoder/config"
	"hstin/gocoder/mapping"
	"hstin/gocoder/structures"
	"log"
	"os"
	"os/exec"

	"github.com/paulmach/orb"
	geojson "github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/tidwall/rtree"
	_ "modernc.org/sqlite"
)

type DataAdminArea struct {
	Placetype string
	Country   string
	Names     map[string]string
	Geometry  orb.Geometry
}

type AdminArea struct {
	Regions  map[string]structures.Region
	Polygons []orb.Polygon
	Country  string
}

type AdminTree struct {
	adminTree   rtree.RTree
	countryTree rtree.RTree
	countryGrid *structures.UniformGridIndex
}

func LoadAdminAreas() AdminTree {

	log.Println("[ADMIN] Loading admin areas")

	db, err := sql.Open("sqlite3", config.WhosOnFirst)
	if err != nil {
		log.Fatalf("[ADMIN] Failed to initialize Who's on First database: %v", err)
	}
	defer db.Close()

	ids, err := db.Query("SELECT geojson.id,spr.placetype,geojson.body,spr.country FROM geojson JOIN spr ON geojson.id = spr.id WHERE spr.placetype NOT IN ('locality', 'empire') GROUP BY geojson.id")
	if err != nil {
		log.Fatalf("[ADMIN] Failed to query GeoJSON features: %v", err)
	}

	adminTrees := AdminTree{}

	stmt, err := db.Prepare("SELECT name, country, language FROM names WHERE id = ? AND privateuse = 'preferred'")
	if err != nil {
		log.Fatalf("[ADMIN] Error preparing statement: %v", err)
	}
	defer stmt.Close()

	loadedAdminAreas := 0

	for ids.Next() {
		var id int64
		var placetype string
		var geojsonBody string
		var country string
		if err := ids.Scan(&id, &placetype, &geojsonBody, &country); err != nil {
			log.Fatalf("[ADMIN] Failed to scan row: %v", err)
		}

		ugeojson, err := geojson.UnmarshalFeature([]byte(geojsonBody))
		if err != nil {
			log.Fatalf("[ADMIN] Failed to unmarshal GeoJSON: %v", err)
		}
		namesMap := make(map[string]string)

		namesMap["placetype"] = placetype
		namesMap["country"] = ugeojson.Properties.MustString("wof:country", "")
		namesMap["name"] = ugeojson.Properties.MustString("name", "")

		var officialLanguages []string

		if langSlice, ok := ugeojson.Properties["wof:lang_x_official"].([]interface{}); ok {
			officialLanguages = make([]string, len(langSlice))
			for i, v := range langSlice {
				if str, ok := v.(string); ok {
					officialLanguages[i] = mapping.Language3ToLanguage2[str]
				}
			}
		} else {
			if defaultLang, ok := mapping.Country3ToLanguage[namesMap["country"]]; ok {
				officialLanguages = []string{mapping.Language3ToLanguage2[defaultLang]}
			} else {
				officialLanguages = []string{"en"}
			}
		}

		names, err := stmt.Query(id)
		if err != nil {
			log.Fatalf("[ADMIN] Error preparing statement: %v", err)
		}
		defer names.Close()

		for names.Next() {
			var name sql.NullString
			var language sql.NullString
			var country sql.NullString
			if err := names.Scan(&name, &country, &language); err != nil {
				log.Fatalf("[ADMIN] Error preparing statement: %v", err)
			}

			if name.Valid && language.Valid {

				namesMap["name:"+mapping.Language3ToLanguage2[language.String]] = name.String

				if ugeojson.Properties["country"] == "" && country.Valid {
					namesMap["country"] = mapping.Iso3ToIso2[country.String]
				}
			}

		}

		for _, lang := range append(officialLanguages, config.Languages...) {
			if namesMap["name"] == "" {
				if name, ok := namesMap["name:"+lang]; ok {
					namesMap["name"] = name
				}
			}
		}

		for _, lang := range config.Languages {
			if namesMap["name:"+lang] == "" {
				namesMap["name:"+lang] = namesMap["name"]
			}
		}

		admin := DataAdminArea{
			Placetype: placetype,
			Names:     namesMap,
			Geometry:  ugeojson.Geometry,
		}

		bounds := ugeojson.Geometry.Bound()

		if admin.Placetype == "region" || admin.Placetype == "county" {
			adminTrees.adminTree.Insert([2]float64{bounds.Min[0], bounds.Min[1]}, [2]float64{bounds.Max[0], bounds.Max[1]}, admin)
		} else if admin.Names["country"] != "" {
			adminTrees.countryTree.Insert([2]float64{bounds.Min[0], bounds.Min[1]}, [2]float64{bounds.Max[0], bounds.Max[1]}, admin)
		}
		loadedAdminAreas++
	}

	log.Printf("[ADMIN] Loaded %d admin areas", loadedAdminAreas)

	adminTrees.countryGrid = GenerateCountries()

	return adminTrees
}

func (a AdminTree) GetCounty(lat, lng float64) AdminArea {
	var adminAreas map[string]DataAdminArea = make(map[string]DataAdminArea)
	var country string = ""

	a.adminTree.Search([2]float64{lng, lat}, [2]float64{lng, lat}, func(min, max [2]float64, data interface{}) bool {
		geojson := data.(DataAdminArea)
		if geojson.Geometry.GeoJSONType() == "Polygon" {
			polygon := geojson.Geometry.(orb.Polygon)
			if planar.PolygonContains(polygon, orb.Point{lng, lat}) {
				adminAreas[geojson.Placetype] = geojson
				if geojson.Names["country"] != "" && country == "" {
					country = geojson.Names["country"]
				}
			}

		} else if geojson.Geometry.GeoJSONType() == "MultiPolygon" {
			multiPolygon := geojson.Geometry.(orb.MultiPolygon)
			for _, polygon := range multiPolygon {
				if planar.PolygonContains(polygon, orb.Point{lng, lat}) {
					adminAreas[geojson.Placetype] = geojson
					if geojson.Names["country"] != "" && country == "" {
						country = geojson.Names["country"]
					}
				}
			}
		}

		_, okRegion := adminAreas["region"]

		_, okCounty := adminAreas["county"]

		if okRegion && okCounty {
			return false
		}

		return true
	})

	if country == "" {
		a.countryTree.Search([2]float64{lng, lat}, [2]float64{lng, lat}, func(min, max [2]float64, data interface{}) bool {
			geojson := data.(DataAdminArea)
			if geojson.Geometry.GeoJSONType() == "Polygon" {
				polygon := geojson.Geometry.(orb.Polygon)
				if planar.PolygonContains(polygon, orb.Point{lng, lat}) {
					if geojson.Names["country"] != "" {
						country = geojson.Names["country"]
						return false
					}
				}
			} else if geojson.Geometry.GeoJSONType() == "MultiPolygon" {
				multiPolygon := geojson.Geometry.(orb.MultiPolygon)
				for _, polygon := range multiPolygon {
					if planar.PolygonContains(polygon, orb.Point{lng, lat}) {
						if geojson.Names["country"] != "" {
							country = geojson.Names["country"]
							return false
						}
					}
				}
			}

			return true
		})
	}

	if country == "" {
		c := a.countryGrid.Search([2]float64{lng, lat})
		if c != nil {
			country = c.Properties["ISO3166-1:alpha2"].(string)
		}
	}

	regions := make(map[string]structures.Region)

	for _, lang := range config.Languages {
		regions[lang] = structures.Region{
			Region:    adminAreas["region"].Names["name:"+lang],
			SubRegion: adminAreas["county"].Names["name:"+lang],
		}
	}

	regions["name"] = structures.Region{
		Region:    adminAreas["region"].Names["name"],
		SubRegion: adminAreas["county"].Names["name"],
	}

	return AdminArea{
		Regions: regions,
		Country: country,
	}
}

func GenerateCountries() *structures.UniformGridIndex {

	log.Println("[COUNTRY] Loading countries")

	exists := false
	_, err := os.Stat(config.Countries)
	if os.IsNotExist(err) {
		exists = false
	} else {
		exists = true
	}

	if exists {
		log.Println("[COUNTRY] Countries already exist, skipping generation...")

		geojsonBytes, err := os.ReadFile(config.Countries)
		if err != nil {
			return nil
		}
		fc, err := geojson.UnmarshalFeatureCollection(geojsonBytes)
		if err != nil {
			return nil
		}

		log.Println("[COUNTRY] Loaded", len(fc.Features), "country features")

		return structures.LoadCountriesWithGrid(fc, 2)
	}

	log.Println("[COUNTRY] Filtering out non-admin boundaries")

	cmd := exec.Command("osmium", "tags-filter", config.Planet, "r/boundary=administrative", "--overwrite", "-o", "boundary.osm")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("[COUNTRY] Error running command: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	log.Println("[COUNTRY] Filtering out non-admin2 boundaries")

	cmd = exec.Command("osmium", "tags-filter", "boundary.osm", "r/admin_level=2", "--overwrite", "-o", "admin2.osm")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("[COUNTRY] Error running command: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	log.Println("[COUNTRY] Exporting to GeoJSON")

	cmd = exec.Command("osmium", "export", "admin2.osm", "-f", "geojson", "--overwrite", "-o", "cc.geojson")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("[COUNTRY] Error running command: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	input, err := os.ReadFile("cc.geojson")
	if err != nil {
		log.Fatalf("[COUNTRY] Error reading input file: %v", err)
	}

	fc, err := geojson.UnmarshalFeatureCollection([]byte(input))
	if err != nil {
		log.Fatalf("[COUNTRY] Error unmarshalling FeatureCollection: %v", err)
	}

	newFC := geojson.NewFeatureCollection()

	loadedCountryFeatures := 0

	log.Println("[COUNTRY] Flattening country features")

	for _, feature := range fc.Features {

		if _, ok := feature.Properties["ISO3166-1:alpha2"]; !ok {
			continue
		}

		switch feature.Geometry.GeoJSONType() {
		case "Polygon":
			newFC.Append(feature)
		case "MultiPolygon":
			for _, polygon := range feature.Geometry.(orb.MultiPolygon) {
				subFeature := geojson.NewFeature(polygon)
				subFeature.Properties = feature.Properties
				newFC.Append(subFeature)
			}
		default:
		}

		loadedCountryFeatures++
	}

	log.Printf("[COUNTRY] Loaded %d country features", loadedCountryFeatures)

	output, err := json.Marshal(newFC)
	if err != nil {
		log.Fatalf("[COUNTRY] Error marshalling new FeatureCollection: %v", err)
	}

	outFile, err := os.Create(config.Countries)
	if err != nil {
		log.Fatalf("[COUNTRY] Error creating output file: %v", err)
	}
	defer outFile.Close()

	_, err = outFile.Write(output)
	if err != nil {
		log.Fatalf("[COUNTRY] Error writing to output file: %v", err)
	}

	// REMOVE INTERMEDIATE FILES
	os.Remove("boundary.osm")
	os.Remove("admin2.osm")
	os.Remove("cc.geojson")

	geojsonBytes, err := os.ReadFile(config.Countries)
	if err != nil {
		return nil
	}
	fCollection, err := geojson.UnmarshalFeatureCollection(geojsonBytes)
	if err != nil {
		return nil
	}

	return structures.LoadCountriesWithGrid(fCollection, 2)
}
