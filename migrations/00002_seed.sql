-- +goose Up
-- Пользователи
INSERT INTO users (id, name, email, role) VALUES
    (1, 'Анна Ковалёва', 'a.kovaleva@agnis.local', 'admin'),
    (2, 'Дмитрий Орлов', 'd.orlov@agnis.local', 'head'),
    (3, 'Мария Белова', 'm.belova@agnis.local', 'manager');

-- Этапы воронки
INSERT INTO stages (id, name, position, outcome) VALUES
    (1, 'Новый лид',       1, 'open'),
    (2, 'В работе',        2, 'open'),
    (3, 'Ожидает ответа',  3, 'open'),
    (4, 'Согласование',    4, 'open'),
    (5, 'Успешно завершено', 5, 'won'),
    (6, 'Отказ',           6, 'lost');

-- Компании 
INSERT INTO companies (id, name, website, phone) VALUES
    (1, 'ООО «Вектор»', 'https://vector.example', '+74951234567'),
    (2, 'ИП Соколов',   '',                       '+79161112233');

-- Контакты 
INSERT INTO contacts (id, first_name, last_name, phone, email, company_id, owner_id) VALUES
    (1, 'Иван',    'Петров',  '+79991234567', 'i.petrov@vector.example', 1,    3),
    (2, 'Ольга',   'Соколова','+79161112233', '',                         2,    3),
    (3, 'Сергей',  'Лебедев', '',             's.lebedev@mail.example',   NULL, 2);

-- Сделки
INSERT INTO deals (id, title, amount, contact_id, stage_id, owner_id) VALUES
    (1, 'Поставка оборудования', 15000000, 1, 2, 3),
    (2, 'Обслуживание сайта',     4500000, 2, 4, 3),
    (3, 'Консультация',                 0, 3, 1, 2);

-- Задачи 
INSERT INTO tasks (id, title, due_at, deal_id, status) VALUES
    (1, 'Отправить КП',          '2026-11-05T12:00:00Z', 1,    'open'),
    (2, 'Позвонить по договору', '2026-11-01T10:00:00Z', 2,    'open'),
    (3, 'Назначить встречу',     '2026-11-03T15:00:00Z', NULL, 'done');


SELECT setval('users_id_seq',    3);
SELECT setval('stages_id_seq',   6);
SELECT setval('companies_id_seq',2);
SELECT setval('contacts_id_seq', 3);
SELECT setval('deals_id_seq',    3);
SELECT setval('tasks_id_seq',    3);

-- +goose Down
TRUNCATE users, stages, companies, contacts, deals, tasks RESTART IDENTITY;