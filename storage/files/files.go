package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"read-adviser-bot/lib/e"
	"read-adviser-bot/storage"
	"time"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	fPath := filepath.Join(s.basePath, page.UserName)

	createFolder, err := folderIsNotExists(fPath)
	if err != nil {
		msg := fmt.Sprintf("can't check if folder %s exists", fPath)
		return e.Wrap(msg, err)
	}

	if createFolder == true {
		if err := os.Mkdir(fPath, defaultPerm); err != nil {
			return err
		}
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)

	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't save", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	//rand.Seed(time.Now().UnixNano())
	rand.NewSource(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePath(filepath.Join(path, file.Name()))

}

func (s Storage) Remove(p *storage.Page) error {
	filename, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, filename)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)
		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {

	filename, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, filename)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", path)
		return false, e.Wrap(msg, err)
	}

	return true, nil

}

func (s Storage) decodePath(filepath string) (*storage.Page, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil

}
func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}

func folderIsNotExists(fPath string) (bool, error) {

	switch _, err := os.Stat(fPath); {
	case errors.Is(err, os.ErrNotExist):
		return true, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if folder %s exists", fPath)
		return false, e.Wrap(msg, err)
	}

	return false, nil
}
