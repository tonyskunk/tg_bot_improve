package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5" //tgbotapi
)

var gBot *tgbotapi.BotAPI
var gToken string
var gChatId int64

var gUsersInChat Users

var gUsefulActivities = Activities{
	// Развитие
	{"sport", "Спорт (15 минут)", 1},
	{"meditation", "Медитация (15 минут)", 1},
	{"language", "Изучение иностранного языка (15 минут)", 1},
	{"swimming", "Плаванье", 4},
	{"walk", "Прогулка (15 минут)", 1},
	{"chores", "Уборка", 2},

	//Работа
	{"work_learning", "Изучение материалов по работе (15 минут)", 1},
	{"portfolio_work", "Работа над проектом для портфолио (15 минут)", 1},
	{"resume_edit", "Редактирование резюме (5 минут)", 1},

	//Творчество
	{"creative", "Творческое созидание (15 минут)", 1},
	{"reading", "Чтение худ. литературы (15 минут)", 1},
}

var gRewards = Activities{
	//Просмотр
	{"watch_series", "Просмотр сериала (1 серия)", 10},
	{"watch_movie", "Просмотр фильма (1 шт)", 20},
	{"social_nets", "Просмотр соц.сетей (30 минут)", 10},

	// Еда
	{"eat_sweets", "Вредное хрючево", 40},
}

type User struct {
	id    int64
	name  string
	coins uint16
}

type Users []*User

type Activity struct {
	code, name string
	coins      uint16
}

type Activities []*Activity

func init() {
	//_ = os.Setenv(TOKEN_NAME_IN_OS, "INSERT_YOUR_TOKEN")
	_ = os.Setenv(TOKEN_NAME_IN_OS, "Your token")

	if gToken = os.Getenv(TOKEN_NAME_IN_OS); gToken == "" {
		panic(fmt.Errorf(`failed to load environment variable "%s"`, TOKEN_NAME_IN_OS))
	}
	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}

	gBot.Debug = true

}

func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func isCallbackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

func delay(seconds uint8) {
	time.Sleep(time.Second * time.Duration(seconds))
}

func printSystemMessageWithDelay(delayInSec uint8, message string) {
	gBot.Send(tgbotapi.NewMessage(gChatId, message))
	delay(delayInSec)
}

func printIntro(update *tgbotapi.Update) {
	printSystemMessageWithDelay(2, "Салют! "+EMOJI_SUNGLASSES)
	printSystemMessageWithDelay(7, "Существует множество полезных действий, регулярно выполняя которые, мы улучшаем качество своей жизни. Однако зачастую веселее, проще или вкуснее сделать что-то вредное. Не так ли?")
	printSystemMessageWithDelay(7, "С большей вероятностью мы предпочтем залипать в Reels вместо урока английского, купить чизборгу во вкусно и дрочка вместо овощей или поваляться вместо того, чтобы заняться физкультурой.")
	printSystemMessageWithDelay(1, EMOJI_SAD)
	printSystemMessageWithDelay(10, "Каждый играл хотя бы в одну игру, где нужно прокачивать персонажа, делая его пизже. Это приятно, потому что каждое действие приносит результат. Однако в реальной жизни систематические действия только со временем начинают становиться заметными. Давай это изменим, лады?")
	printSystemMessageWithDelay(1, EMOJI_SMILE)
	printSystemMessageWithDelay(14, `Перед тобой две таблицы: "Полезные занятия" и "Награды". В первой таблице перечислены простые короткие действия, за выполнение каждого из которых ты заработаешь указанное количество монет. Во второй таблице ты увидишь список действий, которые ты сможешь выполнять только после оплаты за них монетами, заработанными на предыдущем шаге.`)
	printSystemMessageWithDelay(1, EMOJI_COIN)
	printSystemMessageWithDelay(10, `Например, вы полчаса занимаетесь спортом, за что получаете 2 монеты. После этого вам предстоит 2 часа изучения английского языка, за которые вы получите 8 монет. Теперь ты можешь глянуть одну серию «Физрука» с Нагиевым. Это просто, мой белый друг!`)
	printSystemMessageWithDelay(6, `Отмечай совершенные полезные активности, чтобы не потерять монеты.И не забывай "купить" вознаграждение!`)
}

func getKeyboardRow(buttonText, buttonCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonCode))

}

func askToPrintIntro() {
	msg := tgbotapi.NewMessage(gChatId, "Во вступительный сообщениях ты можешь найти смысл данного бота и правила игры. Что думаешь?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_PRINT_INTRO, BUTTON_CODE_PRINT_INTRO),
		getKeyboardRow(BUTTON_TEXT_SKIP_INTRO, BUTTON_CODE_SKIP_INTRO),
	)
	gBot.Send(msg)
}

func showMenu() {
	msg := tgbotapi.NewMessage(gChatId, "Выбери один из вариантов:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_BALANCE, BUTTON_CODE_BALANCE),
		getKeyboardRow(BUTTON_TEXT_USEFUL_ACTIVITIES, BUTTON_CODE_USEFUL_ACTIVITIES),
		getKeyboardRow(BUTTON_TEXT_REWARDS, BUTTON_CODE_REWARDS),
	)
	gBot.Send(msg)
}

func showBalance(user *User) {
	msg := fmt.Sprintf("%s, твой кошелек пока пуст %s \nЗатрекай полезное действие чтобы получить монеты", user.name, EMOJI_DONT_KNOW)
	if coins := user.coins; coins > 0 {
		msg = fmt.Sprintf("%s, у тебя %d %s", user.name, coins, EMOJI_COIN)
	}
	gBot.Send(tgbotapi.NewMessage(gChatId, msg))

	showMenu()

}

func callbackQueryIsMissing(update *tgbotapi.Update) bool {
	return update.CallbackQuery == nil || update.CallbackQuery.From == nil
}

func getUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryIsMissing(update) {
		return
	}

	userId := update.CallbackQuery.From.ID
	for _, userInChat := range gUsersInChat {
		if userId == userInChat.id {
			return userInChat, true
		}
	}
	return
}

func storeUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryIsMissing(update) {
		return
	}
	from := update.CallbackQuery.From
	user = &User{id: from.ID, name: strings.TrimSpace(from.FirstName + " " + from.LastName), coins: 0}
	gUsersInChat = append(gUsersInChat, user)
	return user, true
}

/*func showActivities() (activities Activities, message string, isUseful bool)(Activities, string, bool) {
	activitiesButtonsRows := make([]([]tgbotapi.InlineKeyboardButton), 0, len(gUsefulActivities)+1)
	for _, activity := range activities {
		activityDescription := ""
		if isUseful {
			activityDescription = fmt.Sprintf("+ %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		} else {
			activityDescription = fmt.Sprintf("- %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		}
		activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(activityDescription, activity.code))
	}
	activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(BUTTON_TEXT_PRINT_MENU, BUTTON_CODE_PRINT_MENU))

	msg := tgbotapi.NewMessage(gChatId, message)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(activitiesButtonsRows...)
	gBot.Send(msg)
}*/

func showActivities(activities Activities, message string, isUseful bool) {
	activitiesButtonsRows := make([]([]tgbotapi.InlineKeyboardButton), 0, len(activities)+1)
	for _, activity := range activities {
		activityDescription := ""
		if isUseful {
			activityDescription = fmt.Sprintf("+ %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		} else {
			activityDescription = fmt.Sprintf("- %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		}
		activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(activityDescription, activity.code))
	}
	activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(BUTTON_TEXT_PRINT_MENU, BUTTON_CODE_PRINT_MENU))

	msg := tgbotapi.NewMessage(gChatId, message)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(activitiesButtonsRows...)
	gBot.Send(msg)

}

// func showUsefulActivities() (Activities, string, bool) {
func showUsefulActivities() {
	showActivities(gUsefulActivities, "Трекай полезное действие или возвращайся в главное меню:", true)
}

func showRewards() {
	showActivities(gRewards, "Приобретите награду или вернитесь в главное меню:", false)
}

func findActivity(activities Activities, choiceCode string) (activity *Activity, found bool) {
	for _, activity := range activities {
		if choiceCode == activity.code {
			return activity, true
		}
	}
	return
}

func processUsefulActivity(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`действие "%s" не имеет указанной стоимости`, activity.name)
	} else if user.coins+activity.coins > MAX_USER_COINS {
		errorMsg = fmt.Sprintf("ты не можешь иметь больше, чем %d %s", MAX_USER_COINS, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, Сорян, но %s %s твой баланс остался неизменным.", user.name, errorMsg, EMOJI_SAD)
	} else {
		user.coins += activity.coins
		resultMessage = fmt.Sprintf(`%s, действие "%s" завершено! %d %s добавлен на твой аккаунт. Так держать! %s%s Теперь у тебя есть %d %s`,
			user.name, activity.name, activity.coins, EMOJI_COIN, EMOJI_BICEPS, EMOJI_SUNGLASSES, user.coins, EMOJI_COIN)
	}
	gBot.Send(tgbotapi.NewMessage(gChatId, resultMessage))
	//sendStringMessage(resultMessage)
}

func processReward(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`награда "%s" не имеет указанной стоимости`, activity.name)
	} else if user.coins < activity.coins {
		errorMsg = fmt.Sprintf(`у тебя сейчас есть %d %s. Вы не можете позволить себе "%s" за %d %s`, user.coins, EMOJI_COIN, activity.name, activity.coins, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, экскузумуа, но %s %s твой баланс остался неизменным, награда недоступна %s", user.name, errorMsg, EMOJI_SAD, EMOJI_DONT_KNOW)
	} else {
		user.coins -= activity.coins
		resultMessage = fmt.Sprintf(`%s, награда "%s" выплачена, приступайте! %d %s было списано с вашего счета. Теперь у вас есть %d %s`, user.name, activity.name, activity.coins, EMOJI_COIN, user.coins, EMOJI_COIN)
	}
	gBot.Send(tgbotapi.NewMessage(gChatId, resultMessage))
	//sendStringMessage(resultMessage)
}

func updateProcessing(update *tgbotapi.Update) {
	user, found := getUserFromUpdate(update)
	if !found {
		if user, found = storeUserFromUpdate(update); !found {
			gBot.Send(tgbotapi.NewMessage(gChatId, "Не получается идентифицировать пользователя"))
			//sendStringMessage("Не получается идентифицировать пользователя")
			return
		}

	}

	choiceCode := update.CallbackQuery.Data
	log.Printf("[%T] %s", time.Now(), choiceCode)

	switch choiceCode {
	case BUTTON_CODE_BALANCE:
		showBalance(user)
	case BUTTON_CODE_USEFUL_ACTIVITIES:
		showUsefulActivities()
	case BUTTON_CODE_REWARDS:
		showRewards()
	case BUTTON_CODE_PRINT_INTRO:
		printIntro(update)
		showMenu()
	case BUTTON_CODE_SKIP_INTRO:
		showMenu()
	case BUTTON_CODE_PRINT_MENU:
		showMenu()
	default:
		if usefulActivity, found := findActivity(gUsefulActivities, choiceCode); found {
			processUsefulActivity(usefulActivity, user)
		}

		delay(2)
		showUsefulActivities()
		return
	}

	if reward, found := findActivity(gRewards, choiceCode); found {
		processReward(reward, user)

		delay(2)
		showRewards()
		return
	}

	log.Printf(`[%T] !!!!!!!!! ERROR: Неизвестный код "%s"`, time.Now(), choiceCode)
	msg := fmt.Sprintf("%s, прости, я не узнаю код '%s' %s Пожалуйста, сообщите об этой ошибке моему создателю.", user.name, choiceCode, EMOJI_SAD)
	gBot.Send(tgbotapi.NewMessage(gChatId, msg))

}

func main() {

	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	for update := range gBot.GetUpdatesChan(updateConfig) {
		if isCallbackQuery(&update) {
			updateProcessing(&update)

		} else if isStartMessage(&update) {

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			gChatId = update.Message.Chat.ID
			askToPrintIntro()
		}

	}
}
