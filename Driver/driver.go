package Driver

import (
	"errors"

	"strings"

	"log"
	"os"

	"github.com/google/uuid"

	"path"
)

type DataBase struct {
	Location string
	//Logger      *log.Logger
	collections Collection
	//FileChan    chan []byte
}

// Done
//
//	func doesFileExist(fileName string) bool {
//		_, error := os.Stat(fileName)
//		return os.IsNotExist(error)
//	}
func NewDB(loc string, logger string, col Collection) *DataBase {
	_, err := os.Stat(loc)
	if os.IsNotExist(err) {
		err = os.Mkdir(loc, 0777)
		if err != nil {
			panic(err)
		}
		// var (
		// 	fs *os.File
		// 	e  error
		// )

		// open log file
		logFile, err := os.OpenFile(logger, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("Logging to custom file")
		CreateCollectionFiles("." + loc)
		return &DataBase{
			Location: loc,
			//FileChan: c,
			collections: col,
		}
	} else {
		logFile, err := os.OpenFile(logger, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("Logging to custom file")
		CreateCollectionFiles("." + loc)
		return &DataBase{
			Location: loc,
			//FileChan: c,
			collections: col,
		}
	}

}

// Done

// Done
func (d *DataBase) CreateCollection(name string) error {
	loc := path.Join(d.Location, name)

	_, err := os.Stat(loc)

	if os.IsNotExist(err) {
		log.Printf("%s collection Created \n", name)
		err = os.Mkdir(loc, 0777)
		d.collections.AddCollection(name)
		log.Println(d.collections)
		d.collections.Commit(d.Location)
		if err != nil {
			return (err)
		}
	} else {
		log.Printf("%s collection already exists \n", name)
	}
	return nil
}

func (d *DataBase) IsCollectionExist(name string) bool {
	loc := path.Join(d.Location, name)
	_, err := os.Stat(loc)
	return !os.IsNotExist(err)
}

// Done
func (d *DataBase) PopulateRecords(collection string, data []byte) (message string, err error) {

	ObjId := createuuid()
	fileName := ObjId + ".json"
	fileLocation := path.Join(d.Location, collection, fileName)
	jsonMap := BuildWrapper(data)
	jsonMap.AddField("id", ObjId)

	file, err := os.Create(fileLocation)

	if err != nil {
		log.Panic(err)
		return "Something went wrong", err
	}
	defer file.Close()
	_, err = file.Write(jsonMap.ToBytes())

	if err != nil {
		log.Panic(err)
		return "Something went wrong", err
	}

	log.Println("Data Addes successfully")

	return "Data Addes successfully", nil
}

// Done
func createuuid() string {
	return uuid.New().String()
}

// Done
func (d *DataBase) ReadAll(collection string, limit int) ([]Wrapper, error) {
	//count := 0
	w := []Wrapper{}
	loc := path.Join(d.Location, collection)
	records, err := os.ReadDir(loc)
	if err != nil {
		return nil, errors.New("Collection Not exixts")
	}
	if limit > (len(records) - 1) {
		log.Println("LIMIT EXITED; LIMIT IS ", len(records))
		limit = len(records) - 1
	}
	resultCh := make(chan Wrapper)
	errorCh := make(chan error)
	for i := 0; i < limit; i++ {
		go func(record os.DirEntry) {
			r, err := os.ReadFile(path.Join(d.Location, collection, record.Name()))
			if err != nil {
				errorCh <- err
				return
			}
			resultCh <- *BuildWrapper(r)
		}(records[i])
	}
	// for _, record := range records {
	// 	// if count == limit {
	// 	// 	break
	// 	// }

	// 	go func(record os.DirEntry) {
	// 		r, err := os.ReadFile(path.Join(d.Location, collection, record.Name()))
	// 		if err != nil {
	// 			errorCh <- err
	// 			return
	// 		}
	// 		resultCh <- *BuildWrapper(r)
	// 	}(record)
	// 	//count += 1
	// }
	for i := 0; i < limit; i++ {
		select {
		case wrapper := <-resultCh:
			w = append(w, wrapper)
		case err := <-errorCh:
			close(resultCh)
			close(errorCh)
			log.Panicln(err)
			return nil, err
		}
	}

	//OLD CODE
	// for _, record := range records {

	// 	r, err := os.ReadFile(path.Join(d.Location, collection, record.Name()))
	// 	if err != nil {
	// 		panic("something went wrong line 98")
	// 	}
	// 	w = append(w, *BuildWrapper(r))
	// }
	return w, nil

}

// Done
func (d *DataBase) FindOneById(collection string, id string) (*Wrapper, string, error) {
	if d.IsCollectionExist(collection) {
		loc := path.Join(d.Location, collection)
		files, err := os.ReadDir(loc)
		if err != nil {
			panic("Err line 113")
		}
		for _, v := range files {
			if strings.Split(v.Name(), ".")[0] == id {
				d, err := os.ReadFile(path.Join(loc, v.Name()))
				if err != nil {
					panic("Error 119")
				}
				//return *BuildWrapper(d), nil
				wrapper := BuildWrapper(d)
				return wrapper, path.Join(loc, v.Name()), nil
			}
		}
		return nil, "", errors.New("no record found")
	} else {
		return nil, "", errors.New("collection not found")
	}
}

func (d *DataBase) UpdateOneById(collection, id, filed string, value interface{}) (*Wrapper, error) {
	w, filePath, err := d.FindOneById(collection, id)
	if err != nil {
		return nil, err
	}
	if w.Value()[filed] == nil {
		return nil, nil
	}
	w.Update(filed, value)
	d.commit(filePath, w)
	return w, nil
}

func (d *DataBase) AddField(collection, id, filed string, value interface{}) (*Wrapper, error) {
	w, filePath, err := d.FindOneById(collection, id)
	if err != nil {
		return nil, err
	}

	w.AddField(filed, value)
	d.commit(filePath, w)
	return w, nil
}

func (d *DataBase) commit(recordpath string, w *Wrapper) {
	os.WriteFile(recordpath, w.ToBytes(), os.ModeAppend)
}

func (d *DataBase) ListCollections() *Wrapper {

	//da, _ := os.ReadFile(COLLECTIONFILESLOC)
	da, _ := os.ReadFile(d.Location + "/collections.json")
	return BuildWrapper(da)
}

func (d *DataBase) Where(collection string, field string, value interface{}) ([]Wrapper, error) {
	var reA []Wrapper
	if d.IsCollectionExist(collection) {
		loc := path.Join(d.Location, collection)
		files, err := os.ReadDir(loc)

		if err != nil {
			panic("Err line 113")
		}
		for _, v := range files {

			d, _ := os.ReadFile(path.Join(loc, v.Name()))
			w := BuildWrapper(d)
			if w.Value()[field] == value {
				reA = append(reA, w.Value())
			}
		}
		return reA, nil
	} else {
		return nil, errors.New("collection not found")
	}

}
