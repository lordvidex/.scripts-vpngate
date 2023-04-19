package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func main() {
	data, err := getData()
	if err != nil {
		log.Fatalln("error occured getting data", err)
	}
	links := getLinks(data)
	if len(links) == 0 {
		log.Println("No task to perform, no link was found")
		return
	}
	for _, link := range links {
		fmt.Println(link)
	}
	err = downloadFiles(links[0])
	if err != nil {
		log.Println("error downloading files: ", err.Error())
	} else {
		fmt.Println("successfully downloaded files")
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
	keyMode := true // true for reading into key, false for reading into value
	for i = i + 1; i < len(linkRunes); i++ {
		switch linkRunes[i] {
		case '=':
			keyMode = false // start reading values
		case '&':
			if len(key) != 0 {
				m[string(key)] = string(value)
			}
			key = key[:0]
			value = value[:0]
			keyMode = true

		default:
			if keyMode {
				key = append(key, linkRunes[i])
			} else {
				value = append(value, linkRunes[i])
			}
		}
	}
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

func downloadFiles(link string) error {
	// get the data
	url := parseLinks(link)
	if len(url) == 0 {
		return errors.New("URL is empty")
	}
	for _, u := range url {
		err := dwn(u)
		if err != nil {
			return err
		}
	}
	return nil
}

// downloads each of the ovpn files
func dwn(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Join(err, errors.New("failed downloading file"))
	}
	defer resp.Body.Close()
	// Create the file
	log.Printf("reading file %s into file %s", url, path.Base(url))
	out, err := os.Create(path.Base(url))
	if err != nil {
		return errors.Join(err, errors.New("failed to create file"))
	}
	defer out.Close()

	// Write the response body to the file
	_, err = io.Copy(out, resp.Body)
	return err
}

// getLinks reads the response passed to it and finds a particular pattern that denotes
// the links that lead to configuration page
func getLinks(data []byte) []string {
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
			str, i = grabLink(i)
			links = append(links, str)
			// TODO: we only need the first for now, also, we get speed
			return links
		}
	}
	return links
}

// getData reads the html response from the vpngate website
func getData() ([]byte, error) {
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
