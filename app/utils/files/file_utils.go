package files

import (
	"fmt"
	"io"
	"io/ioutil"
	"k8s-management-go/app/utils/logger"
	"k8s-management-go/app/utils/loggingstate"
	"os"
	"strings"
)

// FileFilter defines possible filter values for a file in the form of prefix and suffix
type FileFilter struct {
	Prefix *string
	Suffix *string
}

// FileOrDirectoryExists checks if file exists
func FileOrDirectoryExists(fileNameWithPath string) bool {
	var log = logger.Log()
	_, err := os.Stat(fileNameWithPath)
	if os.IsNotExist(err) {
		log.Infof("[File Utils] Unable to find file [%s]", fileNameWithPath)
		return false
	}
	return true
}

// ListFilesOfDirectory lists files of a directory if it exists
func ListFilesOfDirectory(directory string) (files *[]string, err error) {
	files, err = ListFilesOfDirectoryWithFilter(directory, nil)
	return files, err
}

// ListFilesOfDirectoryWithFilter lists files of a directory if it exists with a filter
func ListFilesOfDirectoryWithFilter(directory string, filter *FileFilter) (files *[]string, err error) {
	var log = logger.Log()
	// check if the directory exists before reading from directory
	directoryExists := FileOrDirectoryExists(directory)
	if directoryExists {
		fileList, err := ioutil.ReadDir(directory)
		if err != nil {
			log.Errorf("[File Utils] Unable to read directory [%s] %s\n", directory, err.Error())
			return files, err
		}

		var filesReturnValue []string
		for _, file := range fileList {
			if filterFilename(file.Name(), filter) {
				filesReturnValue = append(filesReturnValue, file.Name())
			}
		}
		return &filesReturnValue, err
	}
	return nil, err
}

// filter by filename and filter
func filterFilename(filename string, filter *FileFilter) bool {
	var fileIsOk = true
	// no filter -> everything is ok
	if filter != nil {
		// filter prefix
		if filter.Prefix != nil {
			// filter prefix if exists
			if !strings.HasPrefix(filename, *filter.Prefix) {
				fileIsOk = false
			}
		}
		if filter.Suffix != nil {
			// filter suffix if exists and file is still ok
			if fileIsOk && !strings.HasSuffix(filename, *filter.Suffix) {
				fileIsOk = false
			}
		}

	}
	return fileIsOk
}

// AppendPath is a helper for adding new paths
func AppendPath(originalPath string, pathExtension string) (extendedPath string) {
	// path extension starts with "./" remove it
	if strings.HasPrefix(pathExtension, "./") {
		pathExtension = strings.TrimPrefix(pathExtension, "./")
	}

	// handle suffix and prefix to create proper path
	if strings.HasSuffix(originalPath, "/") {
		if strings.HasPrefix(pathExtension, "/") {
			// originalPath ends with "/" and path extension starts with "/"
			extendedPath = strings.TrimSuffix(originalPath, "/") + pathExtension
		} else {
			// original path ends with "/" and path extension does not start with "/"
			extendedPath = originalPath + pathExtension
		}
	} else if strings.HasPrefix(pathExtension, "/") {
		// original path does not end with "/" but pathExtension has "/" prefix
		extendedPath = originalPath + pathExtension
	} else {
		// original path does not end with "/" and path extension does not start with "/"
		extendedPath = originalPath + "/" + pathExtension
	}
	return extendedPath
}

// CopyFile copies file from src to destination
func CopyFile(src string, dst string) (bytesWritten int64, err error) {
	srcFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !srcFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%v is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	nBytes, err := io.Copy(dstFile, srcFile)

	return nBytes, err
}

// LoadTemplateFilesOfDirectory : Load all template files of a directory
func LoadTemplateFilesOfDirectory(directory string) ([]string, error) {
	var configFiles = ".yaml"
	var fileFilter = FileFilter{
		Suffix: &configFiles,
	}
	filesInDirectory, err := ListFilesOfDirectoryWithFilter(directory, &fileFilter)
	if filesInDirectory == nil {
		loggingstate.AddErrorEntry("-> Could not find any yaml files in directory [" + directory + "].")
		return []string{}, err
	}

	var templateFiles []string
	for _, file := range *filesInDirectory {
		templateFiles = append(templateFiles, AppendPath(directory, file))
	}
	return templateFiles, nil
}

// ReplaceStringInFile replaces content in file
func ReplaceStringInFile(filePath string, stringToReplace string, newString string) (success bool, err error) {
	var log = logger.Log()

	// read file
	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("[ReplaceStringInFile] Cannot read file [%s] \n%s", filePath, err.Error())
		return false, err
	}

	// replace content
	var newContents = strings.Replace(string(read), stringToReplace, newString, -1)

	// write changes
	err = ioutil.WriteFile(filePath, []byte(newContents), 0)
	if err != nil {
		loggingstate.AddErrorEntryAndDetails(fmt.Sprintf("  -> Cannot write file [%s]", filePath), err.Error())
		log.Errorf("[ReplaceStringInFile] Cannot write file [%s] \n%s", filePath, err.Error())
		return false, err
	}
	return true, err
}
