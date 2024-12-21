package utils

import (
	"context"
	"fmt"
	"hash/crc32"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/etsune/bkors/server/models"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func ConvertEntryToMultilineTxt(e models.DBEntry) string {
	isRev := "-"
	if e.IsReviewed {
		isRev = "+"
	}
	dataStr := fmt.Sprintf("%s|%d|%d|%d|%d|%s|%s|%s", isRev, e.Placement.Volume, e.Placement.Page, e.Placement.Side, e.Placement.Paragraph, e.Placement.Coords, e.Id.Hex(), e.Image)
	header := fmt.Sprintf("%s|%s|%s|%s", clearVline(e.Entry.Hangul), clearVline(e.Entry.Hanja), clearVline(e.Entry.HomonymicNumber), clearVline(e.Entry.Transcription))
	body := strings.ReplaceAll(e.Entry.Body, "\n\n", "\n")
	res := fmt.Sprintf("%s\n%s\n%s", dataStr, header, body)
	return res
}

func GetAnonymName(name string) string {
	crc := crc32.ChecksumIEEE([]byte(name))
	str := strconv.FormatUint(uint64(crc), 10) + "12345"
	return "Аноним#" + str[:4]
}

func CompareEdits(src, res string) string {
	dmp := diffmatchpatch.New()
	src = html.EscapeString(src)
	res = html.EscapeString(res)
	diffs := dmp.DiffMain(src, res, false)

	restxt := ""
	for _, diff := range diffs {
		class := ""
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			class = "bg-emerald-200"
		case diffmatchpatch.DiffDelete:
			class = "bg-red-200"
		default:
			class = ""
		}
		restxt += fmt.Sprintf("<span class=\"%s\">%s</span>", class, diff.Text)
	}
	return restxt
}

func ConvertBodyToContent(input string) []*Content {
	// Regular expression to match tags
	tagPattern := regexp.MustCompile(`(\*\*)(.+?)\*\*|(__)(.+?)__`)

	// Split the input based on tag matches
	matches := tagPattern.FindAllStringSubmatchIndex(input, -1)

	var result []*Content
	prevEnd := 0

	for _, match := range matches {
		// Extract text before the current tag
		if match[0] > prevEnd {
			result = append(result, &Content{
				Text: input[prevEnd:match[0]],
			})
		}

		if match[2] == -1 && match[4] == -1 {
			match[2] = match[6]
			match[3] = match[7]
			match[4] = match[8]
			match[5] = match[9]
		}

		// Extract the tag and content
		tag := input[match[2]:match[3]]
		innerContent := input[match[4]:match[5]]

		switch tag {
		case "**":
			tag = "i"
		case "__":
			tag = "a"
		}

		// Parse the inner content recursively
		contentNode := ConvertBodyToContent(innerContent)

		node := &Content{
			Tag:  tag,
		}

		if len(contentNode) == 1 {
			node.Text = contentNode[0].Text
		} else if len(contentNode) > 1 {
			node.Content = contentNode
		} else {
			node.Text = innerContent
		}
		
		result = append(result, node)

		// Update the end of the last processed tag
		prevEnd = match[1]
	}

	// Add remaining text after the last tag
	if prevEnd < len(input) {
		result = append(result, &Content{
			Text: input[prevEnd:],
		})
	}

	return result
}

func CompareEditsAsContent(src, res string) []*Content {
	dmp := diffmatchpatch.New()
	src = html.EscapeString(src)
	res = html.EscapeString(res)
	diffs := dmp.DiffMain(src, res, false)

	resContent := make([]*Content, 0)
	for _, diff := range diffs {
		class := ""
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			class = "bg-emerald-200"
		case diffmatchpatch.DiffDelete:
			class = "bg-red-200"
		default:
			class = ""
		}

		resContent = append(resContent, &Content{
			Class: class,
			Text: diff.Text,
		})
	}
	return resContent
}

type Content struct {
	Tag string
	Class string
	Text string
	Content []*Content
}

func ConvertTime(tc time.Time) string {
	return tc.Format("2006-01-02 15:04:05")
}

func ConvertEditToText(edit models.EditEntry) string {
	isRev := "нет"
	if edit.IsReviewed {
		isRev = "да"
	}

	body := strings.ReplaceAll(edit.Body, "\n", "↵\n")
	body = strings.ReplaceAll(body, "\t", "↹\t")

	text := fmt.Sprintf("%s (%s) [%s] %s\n%s\nОтредактировано: %s", edit.Hangul, edit.Hanja, edit.Transcription, edit.HomonymicNumber, body, isRev)

	return text
}

func GetEditGuide() string {
	resp, err := http.Get("https://raw.githubusercontent.com/etsune/bkors/main/edit_guide.md")
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(data)

	//extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	//p := parser.NewWithExtensions(extensions)
	//doc := p.Parse(data)
	//
	//htmlFlags := html.CommonFlags | html.HrefTargetBlank
	//opts := html.RendererOptions{Flags: htmlFlags}
	//renderer := html.NewRenderer(opts)
	//
	//return string(markdown.Render(doc, renderer))
}

func DangerouslyIncludeHTML(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, s)
		return err
	})
}

func GetMeta(pl *models.Placement) string {
	return fmt.Sprintf("%d|%d|%d|%d|%s", pl.Volume, pl.Page, pl.Side, pl.Paragraph, pl.Coords)
}

func clearVline(txt string) string {
	return strings.ReplaceAll(txt, "|", "")
}

func GetRectColor(isReviewed bool) string {
	if isReviewed {
		return "green"
	}
	return "red"
}

func GetCoords(coord string, pos int) string {
	c := strings.Split(coord, ",")

	switch pos {
	case 1:
		return c[0]
	case 2:
		return c[1]
	case 3:
		xn, _ := strconv.Atoi(c[0])
		x2n, _ := strconv.Atoi(c[2])
		return strconv.Itoa(x2n - xn)
	case 4:
		yn, _ := strconv.Atoi(c[1])
		y2n, _ := strconv.Atoi(c[3])
		return strconv.Itoa(y2n - yn)
	}
	return ""
}

func ConvertEntryToSinglelineTxt(e models.DBEntry) string {
	isRev := "-"
	if e.IsReviewed {
		isRev = "+"
	}

	res := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", e.Entry.Hangul, e.Entry.Hanja, e.Entry.HomonymicNumber, e.Entry.Transcription, e.Entry.Body, isRev, e.Id.Hex())

	return res
}

func GetPrevText() string {
	res := "# Строки, начинающиеся с #, игнорируются\n"
	res += "# \n"
	res += "# Формат статьи:\n"
	res += "# \n"
	res += "# Служебная строка: статья проверена? +/-|том|страница|сторона|параграф|координаты|ID|изображение\n"
	res += "# Заголовок: хангыль|ханча|номер омографа|транскрипция\n"
	res += "# Тело статьи\n\n"

	return res
}

func ParseDataline(line string, expectedLen int) ([]string, error) {
	split := strings.Split(line, "|")
	if len(split) != expectedLen {
		return split, fmt.Errorf("incorrect data line")
	}
	return split, nil
}

func ConvStrToInt(num string) int {
	i, err := strconv.Atoi(num)
	if err != nil {
		i = 0
	}
	return i
}

func GetSortNum(pl models.Placement) int {
	return pl.Volume*10000000 + pl.Page*10000 + pl.Side*1000 + pl.Paragraph
}
