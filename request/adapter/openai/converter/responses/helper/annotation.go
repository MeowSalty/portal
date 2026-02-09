package helper

import (
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertAnnotationsToContract 将 OpenAI Responses 注释转换为统一注释。
func ConvertAnnotationsToContract(annotations []responsesTypes.Annotation) []types.ResponseAnnotation {
	result := make([]types.ResponseAnnotation, 0, len(annotations))

	for _, ann := range annotations {
		contractAnn := types.ResponseAnnotation{
			Extras: make(map[string]interface{}),
		}

		switch {
		case ann.FileCitation != nil:
			v := ann.FileCitation
			contractAnn.Type = "file_citation"
			contractAnn.FileID = &v.FileID
			contractAnn.Extras["openai.responses.index"] = v.Index
			contractAnn.Extras["openai.responses.filename"] = v.Filename

		case ann.URLCitation != nil:
			v := ann.URLCitation
			contractAnn.Type = "url_citation"
			contractAnn.URL = &v.URL
			contractAnn.Title = &v.Title
			contractAnn.StartIndex = &v.StartIndex
			contractAnn.EndIndex = &v.EndIndex

		case ann.ContainerFileCitation != nil:
			v := ann.ContainerFileCitation
			contractAnn.Type = "container_file_citation"
			contractAnn.FileID = &v.FileID
			contractAnn.StartIndex = &v.StartIndex
			contractAnn.EndIndex = &v.EndIndex
			contractAnn.Extras["openai.responses.container_id"] = v.ContainerID
			contractAnn.Extras["openai.responses.filename"] = v.Filename

		case ann.FilePath != nil:
			v := ann.FilePath
			contractAnn.Type = "file_path"
			contractAnn.FileID = &v.FileID
			contractAnn.Extras["openai.responses.index"] = v.Index
		}

		result = append(result, contractAnn)
	}

	return result
}

// ConvertAnnotationsFromContract 将统一注释转换为 OpenAI Responses 注释。
func ConvertAnnotationsFromContract(annotations []types.ResponseAnnotation) []responsesTypes.Annotation {
	result := make([]responsesTypes.Annotation, 0, len(annotations))

	for _, ann := range annotations {
		var annValue responsesTypes.Annotation

		switch ann.Type {
		case "url_citation":
			urlAnn := &responsesTypes.URLCitationAnnotation{}
			if ann.URL != nil {
				urlAnn.URL = *ann.URL
			}
			if ann.Title != nil {
				urlAnn.Title = *ann.Title
			}
			if ann.StartIndex != nil {
				urlAnn.StartIndex = *ann.StartIndex
			}
			if ann.EndIndex != nil {
				urlAnn.EndIndex = *ann.EndIndex
			}
			annValue.URLCitation = urlAnn

		case "file_citation":
			fileAnn := &responsesTypes.FileCitationAnnotation{}
			if ann.FileID != nil {
				fileAnn.FileID = *ann.FileID
			}
			if index, ok := ann.Extras["openai.responses.index"].(int); ok {
				fileAnn.Index = index
			}
			if filename, ok := ann.Extras["openai.responses.filename"].(string); ok {
				fileAnn.Filename = filename
			}
			annValue.FileCitation = fileAnn

		case "container_file_citation":
			containerAnn := &responsesTypes.ContainerFileCitationAnnotation{}
			if ann.FileID != nil {
				containerAnn.FileID = *ann.FileID
			}
			if ann.StartIndex != nil {
				containerAnn.StartIndex = *ann.StartIndex
			}
			if ann.EndIndex != nil {
				containerAnn.EndIndex = *ann.EndIndex
			}
			if containerID, ok := ann.Extras["openai.responses.container_id"].(string); ok {
				containerAnn.ContainerID = containerID
			}
			if filename, ok := ann.Extras["openai.responses.filename"].(string); ok {
				containerAnn.Filename = filename
			}
			annValue.ContainerFileCitation = containerAnn

		case "file_path":
			pathAnn := &responsesTypes.FilePathAnnotation{}
			if ann.FileID != nil {
				pathAnn.FileID = *ann.FileID
			}
			if index, ok := ann.Extras["openai.responses.index"].(int); ok {
				pathAnn.Index = index
			}
			annValue.FilePath = pathAnn
		}

		// 只有设置了至少一个类型才添加到结果
		if annValue.FileCitation != nil || annValue.URLCitation != nil ||
			annValue.ContainerFileCitation != nil || annValue.FilePath != nil {
			result = append(result, annValue)
		}
	}

	return result
}
