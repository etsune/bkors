package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/etsune/bkors/server/models"
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
