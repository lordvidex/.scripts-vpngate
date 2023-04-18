package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	// https://www.vpngate.net/common/openvpn_download.aspx?sid=1681857232096&tcp=1&host=public-vpn-234.opengw.net&port=443&hid=15134981&/vpngate_public-vpn-234.opengw.net_tcp_443.ovpn
	for _, link := range links {
		fmt.Println(link)
	}
	downloadFile(links[0])
}

// cleanLink converts the scraped link to download links
func cleanLink(link string) string {
	tcpString := "tcp"
	if !isTCP {
		tcpString = "udp"
	}
	return fmt.Sprintf("https://www.vpngate.net/common/openvpn_download.aspx?sid=%s&tcp=%d&host=%s&port=%d&hid=%s&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn&/%s_%s_%s.ovpn",
		sid, isTCP, host, port, hid, host, tcpString, port,
	)
}

func downloadFile(link string) error {
	// get the data
	resp, err := http.Get(cleanLink(link))
	if err != nil {
		return errors.Join(err, errors.New("failed downloading file"))
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(filename)
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
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("failed to read day from json body")
	}
	res.Body.Close()
	return data, nil

}
