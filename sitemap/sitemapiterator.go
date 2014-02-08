// sitemapiterator.go
package sitemap

import (
	"bufio"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"time"

	"fmt"
)

type ItemIterator interface {
	// Iterates over the next item in the stream, returns false on error or end of stream.
	// False return value and nil Err() indicate end of stream.
	Next() bool
	// Returns the error caused by the last call to Next()
	Err() error
	//Iterate() chan *SitemapURL
	// Resets the iterator to the start of the stream and causes Item() and Err() to return nil,
	// returns an error when reset fails
	Reset() error
}

type SitemapIndexIterator interface {
	ItemIterator
	// Return the item generated by the last call to Next()
	Item() *SitemapIndexURL
}

type xmlSitemapIndexIterator struct {
	sitemap_decoder *xml.Decoder
	reader          io.ReadSeeker
	err             error
	item            *SitemapIndexURL
	check_empty_xml bool
}

func newXmlSitemapIndexIterator(reader io.ReadSeeker) (iterator *xmlSitemapIndexIterator, err error) {
	if reader == nil {
		err = errors.New("sitemapiterator: reader cannot be nil")
		return nil, err
	}
	iterator = &xmlSitemapIndexIterator{reader: reader, sitemap_decoder: xml.NewDecoder(reader),
		check_empty_xml: true}

	err = validateNewIterator(iterator)
	if err != nil {
		return nil, err
	}
	return iterator, nil
}

func (iterator *xmlSitemapIndexIterator) Reset() (err error) {
	_, err = iterator.reader.Seek(0, os.SEEK_SET)
	iterator.sitemap_decoder = xml.NewDecoder(iterator.reader)
	iterator.check_empty_xml = true
	iterator.err = nil
	iterator.item = nil
	return err
}
func (iterator *xmlSitemapIndexIterator) Err() (err error) {
	return iterator.err
}
func (iterator *xmlSitemapIndexIterator) Item() (item *SitemapIndexURL) {
	return iterator.item
}

func (iterator *xmlSitemapIndexIterator) Next() (success bool) {
	for {
		xml_token, err := iterator.sitemap_decoder.Token()
		if xml_token == nil || err != nil {
			iterator.item = nil
			if err == io.EOF {
				err = nil
			}
			if iterator.check_empty_xml && err == nil {
				err = errors.New(fmt.Sprintf("sitemap: xml file was empty of tags."))
			}
			iterator.err = err
			return false
		}

		switch xml_element := xml_token.(type) {
		case xml.StartElement:
			if xml_element.Name.Local == "sitemap" {
				var xml_sitemap_index_url sitemap_index_entry
				iterator.sitemap_decoder.DecodeElement(&xml_sitemap_index_url, &xml_element)

				sitemap_index_url, err := xml_sitemap_index_url.parseSitemapIndexURL()

				iterator.check_empty_xml = false
				if err != nil && err.UrlParseFailed() {
					iterator.item = nil
					iterator.err = err
					return false
				}

				iterator.item = sitemap_index_url
				iterator.err = err
				return true
			}
		}
	}
	return false
}

type SitemapPageIterator interface {
	ItemIterator
	// Return the item generated by the last call to Next()
	Item() *SitemapURL
}

// Iterator over an xml format sitemap file
type xmlSitemapPageIterator struct {
	sitemap_decoder *xml.Decoder
	reader          io.ReadSeeker
	err             error
	item            *SitemapURL
	check_empty_xml bool
}

func newXmlSitemapPageIterator(reader io.ReadSeeker) (iterator *xmlSitemapPageIterator, err error) {
	if reader == nil {
		err = errors.New("sitemapiterator: reader cannot be nil")
		return nil, err
	}
	iterator = &xmlSitemapPageIterator{reader: reader, sitemap_decoder: xml.NewDecoder(reader),
		check_empty_xml: true}

	err = validateNewIterator(iterator)
	if err != nil {
		return nil, err
	}
	return iterator, nil

}

func (iterator *xmlSitemapPageIterator) Reset() (err error) {
	_, err = iterator.reader.Seek(0, os.SEEK_SET)
	iterator.sitemap_decoder = xml.NewDecoder(iterator.reader)
	iterator.check_empty_xml = true
	iterator.err = nil
	iterator.item = nil
	return err
}
func (iterator *xmlSitemapPageIterator) Err() (err error) {
	return iterator.err
}
func (iterator *xmlSitemapPageIterator) Item() (item *SitemapURL) {
	return iterator.item
}

func (iterator *xmlSitemapPageIterator) Next() (success bool) {
	for {
		xml_token, err := iterator.sitemap_decoder.Token()
		if xml_token == nil || err != nil {
			iterator.item = nil
			if err == io.EOF {
				err = nil
			}
			if iterator.check_empty_xml && err == nil {
				err = errors.New(fmt.Sprintf("sitemap: xml file was empty of tags."))
			}
			iterator.err = err
			return false
		}

		switch xml_element := xml_token.(type) {
		case xml.StartElement:
			if xml_element.Name.Local == "url" {
				var xml_sitemap_url sitemap_url_entry
				iterator.sitemap_decoder.DecodeElement(&xml_sitemap_url, &xml_element)

				sitemap_url, err := xml_sitemap_url.parseSitemapURL()

				iterator.check_empty_xml = false
				if err != nil && err.UrlParseFailed() {
					iterator.item = nil
					iterator.err = err
					return false
				}

				iterator.item = sitemap_url
				iterator.err = err
				return true
			}
		}
	}
	return false
}

// Iterator over a plain text sitemap file
type plainSitemapPageIterator struct {
	sitemap_scanner *bufio.Scanner
	reader          io.ReadSeeker
	err             error
	item            *SitemapURL
}

func newPlainSitemapPageIterator(reader io.ReadSeeker) (iterator *plainSitemapPageIterator, err error) {
	if reader == nil {
		err = errors.New("sitemap: reader cannot be nil")
		return nil, err
	}

	iterator = &plainSitemapPageIterator{reader: reader, sitemap_scanner: bufio.NewScanner(reader)}

	err = validateNewIterator(iterator)
	if err != nil {
		return nil, err
	}
	return iterator, nil
}

func (iterator *plainSitemapPageIterator) Reset() (err error) {
	_, err = iterator.reader.Seek(0, os.SEEK_SET)
	iterator.sitemap_scanner = bufio.NewScanner(iterator.reader)
	iterator.err = nil
	iterator.item = nil
	return err
}

func (iterator *plainSitemapPageIterator) Err() (err error) {
	return iterator.err
}

func (iterator *plainSitemapPageIterator) Item() (item *SitemapURL) {
	return iterator.item
}

func (iterator *plainSitemapPageIterator) Next() (success bool) {
	if !iterator.sitemap_scanner.Scan() {
		iterator.err = iterator.sitemap_scanner.Err()
		iterator.item = nil
		return false
	}
	parsed_url, err := parseURL(iterator.sitemap_scanner.Text())
	if err != nil {
		iterator.err = err
		iterator.item = nil
		return false
	}
	sitemap_url := &SitemapURL{location: parsed_url, last_modified: time.Time{}, change_frequency: "", priority: 1.0}
	iterator.item = sitemap_url
	return true
}

// Validates that the iterator has no errors
func validateNewIterator(iterator ItemIterator) (err error) {
	if !iterator.Next() && iterator.Err() != nil {
		err = iterator.Err()
		iterator.Reset()
		return err
	}
	err = iterator.Reset()
	if err != nil {
		return err
	}
	return nil
}
