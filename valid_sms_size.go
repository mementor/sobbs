package main

import (
	"fmt"
	"math"
	"unicode"
	"unicode/utf8"
)

// RunesType - тип рун в тексте
type RunesType int

const (
	allASCII RunesType = iota
	hasNotASCII

	// Данные взяты из https://spark.ru/startup/targetsms/blog/14089/skolko-simvolov-v-1-sms-soobschenii
	asciiOneSmsSize  = 160
	asciiManySmsSize = 153

	hasNotASCIIOneSmsSize  = 70
	hasNotASCIIManySmsSize = 67
)

// ValidSmsSize проверяет соответствие реального размера СМС предполагаемому
func ValidSmsSize(text string, presumeSize int) error {

	size := GetSmsSize(text)

	if presumeSize != size {
		return fmt.Errorf("presume %d current %d", presumeSize, size)
	}

	return nil
}

// GetSmsSize возвращает размер СМС для заданного текста
func GetSmsSize(text string) int {

	runesCount := utf8.RuneCountInString(text)

	smsType := GetSmsType(text)

	var sizeOneSms int
	var sizeManySms int

	switch smsType {
	case allASCII:
		sizeOneSms = asciiOneSmsSize
		sizeManySms = asciiManySmsSize
	case hasNotASCII:
		sizeOneSms = hasNotASCIIOneSmsSize
		sizeManySms = hasNotASCIIManySmsSize
	}

	if runesCount < sizeOneSms {
		return 1
	}

	size := int(math.Ceil(float64(runesCount) / float64(sizeManySms)))

	return size
}

// GetSmsType возвращает тип рун в тексте. (hasNotASCII,allASCII)
func GetSmsType(text string) RunesType {

	for _, value := range text {
		if value > unicode.MaxASCII {
			return hasNotASCII
		}
	}

	return allASCII
}
