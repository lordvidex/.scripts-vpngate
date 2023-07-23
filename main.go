package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/lordvidex/vpn-gate/osascripts"
)

func main() {
	removeOldConfigs := flag.Bool("rm", false, "removes old configurations before fetchin new ovpn files")
	exceptRemove := flag.String("no-rm", "", "specify the name of the configuration to not remove")
	install := flag.Bool("install", false, "installs downloaded configurations after downloading them")
	clearConfigs := flag.Bool("clear", false, "removes all configurations installed with vpn-gate")
	printConfigs := flag.Bool("print", false, "prints all the current configurations and ends")

	flag.Parse()

	exempted := strings.Split(*exceptRemove, ",")
	if *clearConfigs {
		osascripts.DeleteInstalledConfigs(exempted...)
		return
	}

	if *printConfigs {
		configs, err := osascripts.GetConfigs()
		if err != nil {
			log.Fatalln("error getting configs: ", err.Error())
		}
		for _, config := range configs {
			fmt.Println(config)
		}
		return
	}

	// download section
	data, err := getHTML()
	if err != nil {
		log.Fatalln("error occured getting data", err)
	}
	links := extractLinks(data)
	if len(links) == 0 {
		log.Println("No task to perform, no link was found")
		return
	}
	for _, link := range links {
		fmt.Println(link)
	}
	files, err := downloadFiles(links[0])
	if err != nil {
		log.Println("error downloading files: ", err.Error())
	} else {
		fmt.Println("successfully downloaded files:", strings.Join(files, ", "))
	}

	if *install {
		osascripts.InstallConfigs(files)
		// for _, file := range files {
		// 	delete(file)
		// }
	}

	if *removeOldConfigs {
		exempted = append(exempted, files...)
		osascripts.DeleteInstalledConfigs(exempted...)
	}
}

// parseLinks converts the scraped link to download links
//
// tcp and/or udp links
func parseLinks(link string) []string {
	m := make(map[string]string) // stores the key-value pair from the scraped link
	key, value := make([]rune, 0), make([]rune, 0)
	i := 0
	linkRunes := append([]rune(link), '&') // prevents checking for last k/v pair

	// skip past the first `?`
	for linkRunes[i] != '?' {
		i++
	}
	isKey := true // true for reading into key, false for reading into value
	for i = i + 1; i < len(linkRunes); i++ {
		switch linkRunes[i] {
		case '=':
			isKey = false // start reading values
		case '&':
			if len(key) != 0 {
				m[string(key)] = string(value)
			}
			key = key[:0]
			value = value[:0]
			isKey = true

		default:
			if isKey {
				key = append(key, linkRunes[i])
			} else {
				value = append(value, linkRunes[i])
			}
		}
	}
	// Various examples of download links look like this:
	// https://www.vpngate.net/common/openvpn_download.aspx?sid=1681857232096&tcp=1&host=public-vpn-234.opengw.net&port=443&hid=15134981&/vpngate_public-vpn-234.opengw.net_tcp_443.ovpn
	// https://www.vpngate.net/common/openvpn_download.aspx?sid=1681859492426&udp=1&host=public-vpn-232.opengw.net&port=1195&hid=15134979&/vpngate_public-vpn-232.opengw.net_udp_1195.ovpn
	// https://www.vpngate.net/common/openvpn_download.aspx?sid=1681860645727&tcp=1&host=public-vpn-257.opengw.net&port=443&hid=15135005&/vpngate_public-vpn-257.opengw.net_tcp_443.ovpn
	// https://www.vpngate.net/en/do_openvpn.aspx?fqdn=vpn355390794.opengw.net&ip=143.189.11.160&tcp=0&udp=1936&sid=1681857232096&hid=9736278
	var ans []string
	addProtoLink := func(proto string) {
		ans = append(
			ans,
			fmt.Sprintf(
				"https://www.vpngate.net/common/openvpn_download.aspx?sid=%s&%s=1&host=%s&port=%s&hid=%s&/%s_%s_%s.ovpn",
				m["sid"],
				proto,
				m["fqdn"],
				m[proto],
				m["hid"],
				m["fqdn"],
				proto,
				m[proto],
			),
		)
	}
	if m["tcp"] != "" {
		addProtoLink("tcp")
	}
	if m["udp"] != "" {
		addProtoLink("udp")
	}
	return ans
}

func downloadFiles(link string) ([]string, error) {
	// get the data
	url := parseLinks(link)
	if len(url) == 0 {
		return nil, errors.New("URL is empty")
	}
	files := make([]string, 0, len(url))
	for _, u := range url {
		file, err := downloadFile(u)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func delete(fileName string) {
	// deletes the file that was downloaded if it exists
	if _, err := os.Stat(fileName); err == nil {
		err = os.Remove(fileName)
		if err != nil {
			log.Println("error deleting file: ", err.Error())
		}
	}
}

// downloads each of the ovpn files.
// returns the name of the file and an error if any
func downloadFile(url string) (fileName string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Join(err, errors.New("failed downloading file"))
	}
	defer resp.Body.Close()
	// Create the file
	fileName = path.Base(url)
	log.Printf("reading file %s into file %s", url, fileName)
	out, err := os.Create(fileName)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to create file"))
	}
	defer out.Close()

	// Write the response body to the file
	_, err = io.Copy(out, resp.Body)
	return fileName, err
}

// extractLinks reads the response passed to it and finds a particular pattern that denotes
// the links that lead to configuration page
// TODO: regexp?
func extractLinks(data []byte) []string {
	links := make([]string, 0)
	pattern := []byte("do_openvpn.aspx?fqdn=")
	var b strings.Builder

	// returns the link and the end index
	grabLink := func(initial int) (str string, end int) {
		b.WriteString("do_openvpn.aspx?fqdn=")
		c := data[initial]
		for c != '\'' {
			b.WriteByte(c)
			initial++
			c = data[initial]
		}
		defer b.Reset()
		return b.String(), initial
	}
	for i := 0; i < len(data); i++ {
		found := true
		for j := 0; j < len(pattern); j++ { // guaranteed that i + j can never be the end of the data slice
			if pattern[j] != data[i] {
				found = false
				break
			}
			i++
		}
		if found {
			var str string
			var newI int
			str, newI = grabLink(i)
			i = newI
			links = append(links, str)
			// TODO: we only need the first for now, also, we get speed
			return links
		}
	}
	return links
}

// getHTML reads the html response from the vpngate website
func getHTML() ([]byte, error) {
	res, err := http.Get("https://www.vpngate.net/en/#LIST")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("status code is not 200")
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("failed to read day from json body")
	}
	res.Body.Close()
	return data, nil
}
