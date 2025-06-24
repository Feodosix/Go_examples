package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type Entry struct {
	Athlete string `json:"athlete"`
	Country string `json:"country"`
	Sport   string `json:"sport"`
	Year    int    `json:"year"`
	Gold    int    `json:"gold"`
	Silver  int    `json:"silver"`
	Bronze  int    `json:"bronze"`
}

func main() {
	port := flag.String("port", "", "port to listen on")
	dataPath := flag.String("data", "", "path to data JSON file")
	flag.Parse()
	if *port == "" || *dataPath == "" {
		fmt.Fprintln(os.Stderr, "port and data are required")
		os.Exit(1)
	}

	file, err := os.Open(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open data file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var entries []Entry
	d := json.NewDecoder(file)
	if err := d.Decode(&entries); err != nil {
		var wrap struct {
			Response []Entry `json:"response"`
		}
		file.Seek(0, 0)
		d = json.NewDecoder(file)
		if err2 := d.Decode(&wrap); err2 != nil {
			fmt.Fprintf(os.Stderr, "failed to parse data file: %v\n", err)
			os.Exit(1)
		}
		entries = wrap.Response
	}

	athleteCountry := make(map[string]string)
	athleteEntries := make(map[string][]Entry)
	sportEntries := make(map[string][]Entry)
	yearEntries := make(map[int][]Entry)
	for _, e := range entries {
		if _, ok := athleteCountry[e.Athlete]; !ok {
			athleteCountry[e.Athlete] = e.Country
		}
		athleteEntries[e.Athlete] = append(athleteEntries[e.Athlete], e)
		sportEntries[e.Sport] = append(sportEntries[e.Sport], e)
		yearEntries[e.Year] = append(yearEntries[e.Year], e)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/athlete-info", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		ents, ok := athleteEntries[name]
		if !ok {
			http.Error(w, fmt.Sprintf("athlete %s not found", name), http.StatusNotFound)
			return
		}
		total := map[string]int{"gold": 0, "silver": 0, "bronze": 0}
		byYear := make(map[int]map[string]int)
		for _, e := range ents {
			total["gold"] += e.Gold
			total["silver"] += e.Silver
			total["bronze"] += e.Bronze
			if _, ok := byYear[e.Year]; !ok {
				byYear[e.Year] = map[string]int{"gold": 0, "silver": 0, "bronze": 0}
			}
			byYear[e.Year]["gold"] += e.Gold
			byYear[e.Year]["silver"] += e.Silver
			byYear[e.Year]["bronze"] += e.Bronze
		}
		total["total"] = total["gold"] + total["silver"] + total["bronze"]

		mb := make(map[string]map[string]int)
		for year, m := range byYear {
			y := strconv.Itoa(year)
			mb[y] = map[string]int{
				"gold":   m["gold"],
				"silver": m["silver"],
				"bronze": m["bronze"],
				"total":  m["gold"] + m["silver"] + m["bronze"],
			}
		}

		resp := map[string]interface{}{
			"athlete":        name,
			"country":        athleteCountry[name],
			"medals":         total,
			"medals_by_year": mb,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/top-athletes-in-sport", func(w http.ResponseWriter, r *http.Request) {
		sport := r.URL.Query().Get("sport")
		if sport == "" {
			http.Error(w, "sport is required", http.StatusBadRequest)
			return
		}
		ents, ok := sportEntries[sport]
		if !ok {
			http.Error(w, fmt.Sprintf("sport '%s' not found", sport), http.StatusNotFound)
			return
		}
		limit := 3
		if ls := r.URL.Query().Get("limit"); ls != "" {
			n, err := strconv.Atoi(ls)
			if err != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = n
		}
		totals := make(map[string][3]int)
		for _, e := range ents {
			a := totals[e.Athlete]
			a[0] += e.Gold
			a[1] += e.Silver
			a[2] += e.Bronze
			totals[e.Athlete] = a
		}
		type rec struct {
			Name   string
			Counts [3]int
		}
		list := make([]rec, 0, len(totals))
		for name, cnts := range totals {
			list = append(list, rec{name, cnts})
		}
		sort.Slice(list, func(i, j int) bool {
			a, b := list[i], list[j]
			if a.Counts[0] != b.Counts[0] {
				return a.Counts[0] > b.Counts[0]
			}
			if a.Counts[1] != b.Counts[1] {
				return a.Counts[1] > b.Counts[1]
			}
			if a.Counts[2] != b.Counts[2] {
				return a.Counts[2] > b.Counts[2]
			}
			return a.Name < b.Name
		})
		if limit < 0 {
			limit = 0
		}
		if limit > len(list) {
			limit = len(list)
		}
		res := make([]map[string]interface{}, 0, limit)
		for _, rec := range list[:limit] {
			byYear := make(map[string]map[string]int)
			tot := map[string]int{"gold": 0, "silver": 0, "bronze": 0}
			for _, e := range ents {
				if e.Athlete != rec.Name {
					continue
				}
				tot["gold"] += e.Gold
				tot["silver"] += e.Silver
				tot["bronze"] += e.Bronze
				y := strconv.Itoa(e.Year)
				m := byYear[y]
				if m == nil {
					m = map[string]int{"gold": 0, "silver": 0, "bronze": 0}
				}
				m["gold"] += e.Gold
				m["silver"] += e.Silver
				m["bronze"] += e.Bronze
				byYear[y] = m
			}
			mb := make(map[string]map[string]int)
			for y, m := range byYear {
				mb[y] = map[string]int{
					"gold":   m["gold"],
					"silver": m["silver"],
					"bronze": m["bronze"],
					"total":  m["gold"] + m["silver"] + m["bronze"],
				}
			}
			obj := map[string]interface{}{
				"athlete":        rec.Name,
				"country":        athleteCountry[rec.Name],
				"medals":         map[string]int{"gold": rec.Counts[0], "silver": rec.Counts[1], "bronze": rec.Counts[2], "total": rec.Counts[0] + rec.Counts[1] + rec.Counts[2]},
				"medals_by_year": mb,
			}
			res = append(res, obj)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/top-countries-in-year", func(w http.ResponseWriter, r *http.Request) {
		yearStr := r.URL.Query().Get("year")
		if yearStr == "" {
			http.Error(w, "year is required", http.StatusBadRequest)
			return
		}
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		ents, ok := yearEntries[y]
		if !ok {
			http.Error(w, fmt.Sprintf("year %d not found", y), http.StatusNotFound)
			return
		}
		limit := 3
		if ls := r.URL.Query().Get("limit"); ls != "" {
			n, err := strconv.Atoi(ls)
			if err != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = n
		}
		cm := make(map[string][3]int)
		for _, e := range ents {
			c := cm[e.Country]
			c[0] += e.Gold
			c[1] += e.Silver
			c[2] += e.Bronze
			cm[e.Country] = c
		}
		type recC struct {
			Country string
			Counts  [3]int
		}
		listC := make([]recC, 0, len(cm))
		for country, cnts := range cm {
			listC = append(listC, recC{country, cnts})
		}
		sort.Slice(listC, func(i, j int) bool {
			a, b := listC[i], listC[j]
			if a.Counts[0] != b.Counts[0] {
				return a.Counts[0] > b.Counts[0]
			}
			if a.Counts[1] != b.Counts[1] {
				return a.Counts[1] > b.Counts[1]
			}
			if a.Counts[2] != b.Counts[2] {
				return a.Counts[2] > b.Counts[2]
			}
			return a.Country < b.Country
		})
		if limit < 0 {
			limit = 0
		}
		if limit > len(listC) {
			limit = len(listC)
		}
		res := make([]map[string]interface{}, 0, limit)
		for _, c := range listC[:limit] {
			res = append(res, map[string]interface{}{
				"country": c.Country,
				"gold":    c.Counts[0],
				"silver":  c.Counts[1],
				"bronze":  c.Counts[2],
				"total":   c.Counts[0] + c.Counts[1] + c.Counts[2],
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	http.ListenAndServe(":"+*port, mux)
}
