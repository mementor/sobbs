package main

import "testing"

func TestGetSmsSize(t *testing.T) {
	testdata := []struct {
		Text        string
		PresumeSize int
	}{
		{
			"Привет, Саня. Этот тест для тебя",
			1,
		},
		{
			"Далеко-далеко за словесными горами в стране, гласных и согласных живут рыбные тексты. Инициал дороге, приставка которой заглавных от всех агенство проектах жаренные переулка переписывается злых! Коварный свой речью послушавшись подпоясал, эта буквоград, использовало обеспечивает безорфографичный, большой всеми курсивных составитель, деревни дал снова последний они! Дороге свое всеми они, назад рукописи, предложения осталось на берегу.",
			7,
		},
		{
			"Далеко-далеко за словесными горами в стране, гласных и согласных живут рыбные тексты. Рот, за пунктуация однажды страна, щеке эта от всех! Щеке, его.",
			3,
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipisicing elit. Vero, animi.",
			1,
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipisicing elit. Atque officiis assumenda necessitatibus dolor aliquid, molestiae sunt officia tempora, unde quo rerum aperiam quos, natus veritatis.",
			2,
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipisicing elit. Amet nam iure, tempora ipsum nesciunt beatae? Molestias, ut atque quos dignissimos. Qui aspernatur incidunt excepturi maxime, aperiam, itaque quia veritatis delectus reiciendis et sit voluptas quis, nulla fuga assumenda, sapiente nemo consequuntur sequi animi?",
			3,
		},
	}

	for _, value := range testdata {
		if size := GetSmsSize(value.Text); size != value.PresumeSize {
			t.Errorf("Размер SMS текста \"%s\" не полученному: \nПолученный: %d\nПредполагаемый: %d", value.Text, size, value.PresumeSize)
		}
	}
}

func TestGetSmsType(t *testing.T) {
	testdata := []struct {
		Text string
		Type RunesType
	}{
		{
			"Привет, Саня. Этот тест для тебя",
			hasNotASCII,
		},
		{
			"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!\"#$%&\\'()*+,-./:;<=>?@[\\]^_`{|}~ \t\n\r\x0b\x0c",
			allASCII,
		},
		{
			// В начале а из русской раскладки
			"а0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!\"#$%&\\'()*+,-./:;<=>?@[\\]^_`{|}~ \t\n\r\x0b\x0c",
			hasNotASCII,
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipisicing elit. Vero, animi.",
			allASCII,
		},
		{
			"Привет, Саня. Этот тест для тебя. Lorem ipsum dolor sit amet, consectetur adipisicing elit. Vero, animi.",
			hasNotASCII,
		},
	}

	for _, value := range testdata {
		if resultType := GetSmsType(value.Text); resultType != value.Type {
			t.Errorf("Предполагаемый тип рун SMS текста \"%s\" не соответствует полученному: \nПолученный: %v\nПредполагаемый: %v", value.Text, resultType, value.Type)
		}
	}
}
