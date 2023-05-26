package filedb

import (
	//	"bytes"

	"context"
	"encoding/gob"
	//	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Pashteto/tgBotTimeSqlite/config"
	"github.com/Pashteto/tgBotTimeSqlite/repos"
)

var ctx, _ = context.WithCancel(context.Background())

const urlTTL1 = time.Second * 1000

type iDShortURL struct {
	ID      string
	LongURL string
}

func ID(m iDShortURL) string {
	return m.ID
}
func URL(m iDShortURL) string {
	return m.LongURL
}

type FWriter interface {
	WriteIDShortURL(idShURL []iDShortURL) error
	Close() error
}
type FReader interface {
	ReadIDShortURL() ([]iDShortURL, error)
	Close() error
}

type fWriter struct {
	file    *os.File
	encoder *gob.Encoder
}

type fReader struct {
	file    *os.File
	decoder *gob.Decoder
}

func NewFWriter(fileName string) (*fWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &fWriter{
		file:    file,
		encoder: gob.NewEncoder(file),
	}, nil
}

func NewFReader(fileName string) (*fReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &fReader{
		file:    file,
		decoder: gob.NewDecoder(file),
	}, nil
}

func (p *fWriter) WriteIDShortURL(idShURL *iDShortURL) error {
	return p.encoder.Encode(idShURL)
}

func (c *fReader) ReadIDShortURL() ([]iDShortURL, error) {
	idShURL := []iDShortURL{}
	if err := c.decoder.Decode(&idShURL); err != nil {
		if err.Error() != "EOF" {
			return nil, err
		}
	}
	return idShURL, nil
}

func (p *fWriter) Close() error {
	return p.file.Close()
}

func (c *fReader) Close() error {
	return c.file.Close()
}
func CreateDirFileDBExists(cfg config.Config) error {
	fjnv := strings.SplitAfter(cfg.FStorPath, "/")
	if len(fjnv) > 0 {
		fjnv = fjnv[:len(fjnv)-1]
		fjnv1 := strings.Join(fjnv, "")
		return os.MkdirAll(fjnv1, 0777)
	}
	return nil
}

func UpdateDBSlice(rdb repos.SetterGetter, cfg config.Config) error {
	fileName := cfg.FStorPath
	reader, err := NewFReader(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer reader.Close()
	readIDShortURLSlice, err := reader.ReadIDShortURL()
	if err != nil {
		return err
	}
	errReadDB := pingRedisDB(rdb)
	if errReadDB != nil {
		return errReadDB
	}
	err = rdb.FlushAllKeys(ctx)
	if err != nil {
		return err
	}
	for i := range readIDShortURLSlice {
		strIDShortURL := readIDShortURLSlice[i]
		err = testFiledURLAndConvert(&strIDShortURL)
		if err != nil {
			return err
		}
		key := strIDShortURL.ID
		value := strIDShortURL.LongURL
		err = rdb.Set(ctx, key, value, urlTTL1)
		if err != nil {
			return err
		}
	}
	return nil
}

func testFiledURLAndConvert(in *iDShortURL) error {
	if in == nil {
		return errors.New("nil filed id")
	}
	if in.ID == "" {
		return errors.New("empty filed id")
	}
	if in.LongURL == "" {
		return errors.New("empty filed url")
	}
	longURL, err := url.Parse(in.LongURL)
	if err != nil {
		return errors.New("unable to parse filed url")
	}
	if !longURL.IsAbs() {
		longURL.Scheme = "http"
	}
	in.LongURL = longURL.String()
	return nil
}

func PostInFileDB(id string, longURL *url.URL, cfg config.Config) error {
	fileName := cfg.FStorPath // + "/URLs.txt"
	writer, err := NewFWriter(fileName)
	if err != nil {
		return err
	}
	defer writer.Close()
	idShURL := &iDShortURL{ID: id, LongURL: longURL.String()}
	if err := writer.encoder.Encode(&idShURL); err != nil {
		return err
	}
	return nil
}

func WriteAll(rdb repos.SetterGetter, cfg config.Config) error {
	fileName := cfg.FStorPath

	writer, err := NewFWriter(fileName)
	if err != nil {
		return err
	}
	defer writer.Close()

	var DBWrite []iDShortURL
	keys, err := rdb.ListAllKeys(ctx)
	if err != nil {
		return err
	}
	for _, key := range keys {
		value, err := rdb.Get(ctx, key)
		if err != nil {
			return err
		}
		DBWrite = append(DBWrite, iDShortURL{ID: key, LongURL: value})
	}
	if err := writer.encoder.Encode(&DBWrite); err != nil {
		return err
	}

	return nil
}

func pingRedisDB(client repos.SetterGetter) error {
	if client == nil {
		return errors.New("no redis db")
	}
	err := client.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
