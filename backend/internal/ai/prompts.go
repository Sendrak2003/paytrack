package ai

// Prompts for extracting structured bank operations from raw statement text.
//
// Design notes:
//   - The system prompt pins the model to a strict JSON contract and forbids
//     prose, so the response is machine-parseable.
//   - We describe each field, its type, and normalization rules (date format,
//     amount as number, INN digits only) to reduce post-processing.
//   - A one-shot example anchors the expected shape and disambiguates the
//     Russian banking vocabulary ("Плательщик", "ИНН", "Назначение платежа").
//   - We instruct the model to skip non-payment lines (balances, fees, headers)
//     to cut false positives.

const SystemPrompt = `Ты — парсер банковских выписок российского digital-агентства.
Тебе на вход даётся сырой текст выписки (как извлечён из PDF). Твоя задача —
извлечь ТОЛЬКО входящие оплаты от клиентов и вернуть их строго в формате JSON.

ПРАВИЛА:
1. Возвращай ТОЛЬКО валидный JSON, без markdown, без пояснений, без текста до или после.
2. Корневой объект: {"operations": [ ... ]}.
3. Каждая операция — объект с полями:
   - "date": дата операции в формате "YYYY-MM-DD";
   - "amount": сумма как ЧИСЛО (без пробелов, валюты и кавычек), например 150000.50;
   - "payer_inn": ИНН плательщика, только цифры (10 или 12 знаков), иначе "";
   - "payer_name": наименование плательщика строкой;
   - "purpose": полное назначение платежа;
   - "invoice_number": номер счёта, если есть в назначении (например "101"), иначе "".
4. Бери только ВХОДЯЩИЕ платежи (поступления от клиентов). Пропускай:
   списания, комиссии банка, остатки, обороты, заголовки, итоги.
5. Если поле не определяется — ставь "" для строк и 0 для чисел, не выдумывай.
6. Если оплат нет — верни {"operations": []}.`

const UserPromptTemplate = `Извлеки оплаты из этой выписки:

---
%s
---

Верни JSON по описанному контракту.`

// FewShotExample is appended to the system prompt to anchor the output shape.
const FewShotExample = `ПРИМЕР.
Вход:
"05.10.2024  Поступление  150 000,00 RUB
Плательщик: ООО «Альфа Медиа», ИНН 7701234567
Назначение: Оплата по счёту №101 за разработку сайта, аванс 50%"

Ожидаемый ответ:
{"operations":[{"date":"2024-10-05","amount":150000.00,"payer_inn":"7701234567","payer_name":"ООО «Альфа Медиа»","purpose":"Оплата по счёту №101 за разработку сайта, аванс 50%","invoice_number":"101"}]}`
