package main

import (
	"fmt"
	"log"
	"payments-dashboard/internal/models"
	"time"

	"gorm.io/gorm"
)

func seedDatabase(db *gorm.DB) {
	var count int64
	db.Model(&models.Client{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping.")
		return
	}

	log.Println("Seeding database...")

	// --- Clients / Legal Entities ---
	// Based on bank statement for ИП Громов Артём Викторович
	// (account 40802810937184056213, АО "ФИН-МОСТ БАНК")
	clients := []models.Client{
		{Name: "ООО «Вебстрой»", INN: "7701456789", OGRN: "1027700000001", BankAccount: "40702810500000000011", ContactPerson: "Николаев С.П."},
		{Name: "ИП Сидорова Е.А.", INN: "503412567890", OGRN: "318503400000012", BankAccount: "40802810200000000022", ContactPerson: "Сидорова Е.А."},
		{Name: "ООО «ДиджиталПро»", INN: "7709123456", OGRN: "1027700000003", BankAccount: "40702810300000000033", ContactPerson: "Кузнецов А.В."},
		{Name: "АО «Медиа Групп»", INN: "7712098765", OGRN: "1027700000004", BankAccount: "40702810400000000044", ContactPerson: "Романова И.Д."},
		{Name: "ООО «ФудСервис»", INN: "7756432109", OGRN: "1027700000005", BankAccount: "40702810500000000055", ContactPerson: "Белов М.Н."},
		{Name: "ООО «ТрансЛог»", INN: "7743219876", OGRN: "1027700000006", BankAccount: "40702810600000000066", ContactPerson: "Орлов Д.В."},
	}
	db.Create(&clients)

	// --- Projects ---
	projects := []models.Project{
		{Name: "Корпоративный сайт «Вебстрой»", ClientID: clients[0].ID, Status: "active"},
		{Name: "Интернет-магазин «Сидорова»", ClientID: clients[1].ID, Status: "active"},
		{Name: "SEO-продвижение «ДиджиталПро»", ClientID: clients[2].ID, Status: "active"},
		{Name: "Рекламная кампания «Медиа Групп»", ClientID: clients[3].ID, Status: "active"},
		{Name: "Разработка мобильного приложения «ФудСервис»", ClientID: clients[4].ID, Status: "paused"},
		{Name: "Контент-маркетинг «ТрансЛог»", ClientID: clients[5].ID, Status: "completed"},
	}
	db.Create(&projects)

	// Helper to make time
	d := func(year, month, day int) time.Time {
		return time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
	}

	// --- Payments ---
	// Reconstructed from bank statement: period 15.07.2026 – 14.08.2026
	payments := []models.Payment{
		// Project 1 — Корпоративный сайт «Вебстрой»
		{ProjectID: projects[0].ID, LegalEntityID: clients[0].ID, PaymentDate: d(2026, 7, 15), Amount: 7100, PaymentPurpose: "Оплата за авторскую молнию золотистую типа, сч.от 5463 от 13.07.2026", ServiceStage: "Разработка", InvoiceNumber: "5463", ContractNumber: "Д-2026-01"},
		{ProjectID: projects[0].ID, LegalEntityID: clients[0].ID, PaymentDate: d(2026, 7, 15), Amount: 14100, PaymentPurpose: "Оплата по счёту №б/н за разработку, НДС не облагается", ServiceStage: "Разработка", InvoiceNumber: "б/н", ContractNumber: "Д-2026-01"},
		{ProjectID: projects[0].ID, LegalEntityID: clients[0].ID, PaymentDate: d(2026, 7, 16), Amount: 9500, PaymentPurpose: "Пополнение карты бизнес, счёт от Тинькофф", ServiceStage: "Разработка", InvoiceNumber: "", ContractNumber: "Д-2026-01"},
		{ProjectID: projects[0].ID, LegalEntityID: clients[0].ID, PaymentDate: d(2026, 7, 21), Amount: 251100, PaymentPurpose: "Оплата по счёту за дизайн и разработку корпоративного портала", ServiceStage: "Дизайн", InvoiceNumber: "1247", ContractNumber: "Д-2026-01"},

		// Project 2 — Интернет-магазин «Сидорова»
		{ProjectID: projects[1].ID, LegalEntityID: clients[1].ID, PaymentDate: d(2026, 7, 17), Amount: 23000, PaymentPurpose: "Оплата за разработку и сопровождение интернет-магазина, сч. от 156 от 15.07.2026", ServiceStage: "Разработка", InvoiceNumber: "156", ContractNumber: "Д-2026-02"},
		{ProjectID: projects[1].ID, LegalEntityID: clients[1].ID, PaymentDate: d(2026, 7, 22), Amount: 26800, PaymentPurpose: "Возмещение транспортных расходов и оплата за хостинг", ServiceStage: "Сопровождение", InvoiceNumber: "157", ContractNumber: "Д-2026-02"},
		{ProjectID: projects[1].ID, LegalEntityID: clients[1].ID, PaymentDate: d(2026, 7, 23), Amount: 15200, PaymentPurpose: "Оплата по счёту за настройку платёжной системы на сайте", ServiceStage: "Разработка", InvoiceNumber: "163", ContractNumber: "Д-2026-02"},

		// Project 3 — SEO-продвижение «ДиджиталПро»
		{ProjectID: projects[2].ID, LegalEntityID: clients[2].ID, PaymentDate: d(2026, 7, 17), Amount: 34400, PaymentPurpose: "Оплата по счёту за SEO продвижение, июль 2026", ServiceStage: "SEO", InvoiceNumber: "320", ContractNumber: "Д-2026-03"},
		{ProjectID: projects[2].ID, LegalEntityID: clients[2].ID, PaymentDate: d(2026, 7, 21), Amount: 47000, PaymentPurpose: "Оплата за оптимизацию и написание статей, 15 материалов", ServiceStage: "Контент", InvoiceNumber: "321", ContractNumber: "Д-2026-03"},
		{ProjectID: projects[2].ID, LegalEntityID: clients[2].ID, PaymentDate: d(2026, 8, 1), Amount: 34400, PaymentPurpose: "Оплата по счёту за SEO продвижение, август 2026", ServiceStage: "SEO", InvoiceNumber: "340", ContractNumber: "Д-2026-03"},

		// Project 4 — Рекламная кампания «Медиа Групп»
		{ProjectID: projects[3].ID, LegalEntityID: clients[3].ID, PaymentDate: d(2026, 7, 17), Amount: 19000, PaymentPurpose: "Оплата за настройку контекстной рекламы Яндекс.Директ", ServiceStage: "Реклама", InvoiceNumber: "445", ContractNumber: "Д-2026-04"},
		{ProjectID: projects[3].ID, LegalEntityID: clients[3].ID, PaymentDate: d(2026, 7, 22), Amount: 9300, PaymentPurpose: "Возврат за переплату по рекламному бюджету", ServiceStage: "Реклама", InvoiceNumber: "446", ContractNumber: "Д-2026-04"},
		{ProjectID: projects[3].ID, LegalEntityID: clients[3].ID, PaymentDate: d(2026, 8, 5), Amount: 55000, PaymentPurpose: "Оплата за ведение рекламных кампаний, август 2026", ServiceStage: "Реклама", InvoiceNumber: "460", ContractNumber: "Д-2026-04"},

		// Project 5 — Мобильное приложение «ФудСервис»
		{ProjectID: projects[4].ID, LegalEntityID: clients[4].ID, PaymentDate: d(2026, 7, 22), Amount: 8200, PaymentPurpose: "Оплата по счёту 8.25225 за создание прототипа мобильного приложения", ServiceStage: "Дизайн", InvoiceNumber: "825", ContractNumber: "Д-2026-05"},
		{ProjectID: projects[4].ID, LegalEntityID: clients[4].ID, PaymentDate: d(2026, 7, 28), Amount: 120000, PaymentPurpose: "Аванс за разработку мобильного приложения iOS/Android", ServiceStage: "Разработка", InvoiceNumber: "830", ContractNumber: "Д-2026-05"},

		// Project 6 — Контент-маркетинг «ТрансЛог»
		{ProjectID: projects[5].ID, LegalEntityID: clients[5].ID, PaymentDate: d(2026, 7, 23), Amount: 91000, PaymentPurpose: "Оплата по договору за ведение и администрирование контент-плана", ServiceStage: "Контент", InvoiceNumber: "610", ContractNumber: "Д-2026-06"},
		{ProjectID: projects[5].ID, LegalEntityID: clients[5].ID, PaymentDate: d(2026, 8, 3), Amount: 45000, PaymentPurpose: "Оплата за написание и публикацию материалов, 10 статей", ServiceStage: "Контент", InvoiceNumber: "615", ContractNumber: "Д-2026-06"},
		{ProjectID: projects[5].ID, LegalEntityID: clients[5].ID, PaymentDate: d(2026, 8, 10), Amount: 18500, PaymentPurpose: "Оплата за дизайн баннеров и инфографики", ServiceStage: "Дизайн", InvoiceNumber: "620", ContractNumber: "Д-2026-06"},
	}
	// Assign deterministic ExternalID per seed payment so the uniqueIndex on
	// external_id is satisfied (no two empty strings) and re-seeding stays idempotent.
	for i := range payments {
		payments[i].ExternalID = fmt.Sprintf("seed-%s-%d-%d", payments[i].InvoiceNumber, payments[i].ProjectID, i)
	}
	db.Create(&payments)

	// --- Acts ---
	sentAt1 := d(2026, 7, 18)
	signedAt1 := d(2026, 7, 25)
	sentAt2 := d(2026, 7, 20)
	sentAt3 := d(2026, 7, 22)
	signedAt3 := d(2026, 7, 28)
	sentAt4 := d(2026, 7, 25)
	signedAt4 := d(2026, 7, 30)
	sentAt5 := d(2026, 7, 20)
	signedAt5 := d(2026, 7, 26)
	sentAt6 := d(2026, 7, 25)
	signedAt6 := d(2026, 8, 1)
	sentAt7 := d(2026, 7, 20)
	signedAt7 := d(2026, 7, 27)
	sentAt8 := d(2026, 7, 25)
	sentAt9 := d(2026, 8, 5)
	signedAt9 := d(2026, 8, 8)

	acts := []models.Act{
		// payments[0] — закрыт (корпоративный сайт, первая оплата)
		{PaymentID: payments[0].ID, IsSent: true, SentAt: &sentAt1, IsSigned: true, SignedAt: &signedAt1, ManagerComment: "Акт подписан, всё ок"},
		// payments[1] — отправлен, не подписан (>14 дней — needs_attention)
		{PaymentID: payments[1].ID, IsSent: true, SentAt: &sentAt2, IsSigned: false, ManagerComment: "Клиент не отвечает, повторный звонок 01.08"},
		// payments[2] — не отправлен (свежий)
		// payments[3] — закрыт (дизайн)
		{PaymentID: payments[3].ID, IsSent: true, SentAt: &sentAt3, IsSigned: true, SignedAt: &signedAt3, ManagerComment: ""},

		// payments[4] — закрыт (интернет-магазин, первая оплата)
		{PaymentID: payments[4].ID, IsSent: true, SentAt: &sentAt4, IsSigned: true, SignedAt: &signedAt4, ManagerComment: ""},
		// payments[5] — не отправлен
		// payments[6] — не отправлен

		// payments[7] — закрыт (SEO июль)
		{PaymentID: payments[7].ID, IsSent: true, SentAt: &sentAt5, IsSigned: true, SignedAt: &signedAt5, ManagerComment: ""},
		// payments[8] — закрыт (контент)
		{PaymentID: payments[8].ID, IsSent: true, SentAt: &sentAt6, IsSigned: true, SignedAt: &signedAt6, ManagerComment: ""},
		// payments[9] — не отправлен (август, свежий)

		// payments[10] — закрыт (реклама, настройка)
		{PaymentID: payments[10].ID, IsSent: true, SentAt: &sentAt7, IsSigned: true, SignedAt: &signedAt7, ManagerComment: ""},
		// payments[11] — возврат, нет акта
		// payments[12] — отправлен, ожидает подписи
		{PaymentID: payments[12].ID, IsSent: true, SentAt: &sentAt8, IsSigned: false, ManagerComment: "Отправлен на подпись клиенту"},

		// payments[13] — не отправлен (мобильное приложение, приостановлен)
		// payments[14] — не отправлен

		// payments[15] — закрыт (контент ТрансЛог)
		{PaymentID: payments[15].ID, IsSent: true, SentAt: &sentAt9, IsSigned: true, SignedAt: &signedAt9, ManagerComment: "Проект завершён, все акты закрыты"},
		// payments[16] — не отправлен
		// payments[17] — не отправлен
	}
	db.Create(&acts)

	log.Println("Seeding complete.")
}
