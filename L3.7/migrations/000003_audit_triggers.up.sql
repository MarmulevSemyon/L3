-- Эта функция автоматически обновляет поле updated_at
-- перед каждым UPDATE в таблице items.
CREATE OR REPLACE FUNCTION set_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Триггер для автоматического обновления updated_at.
-- Срабатывает ДО обновления строки товара.
CREATE TRIGGER items_set_updated_at_trigger
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION set_items_updated_at();


-- Эта функция логирует создание товара.
-- OLD здесь нет, потому что строка только создаётся.
-- NEW содержит новую строку товара.
CREATE OR REPLACE FUNCTION log_item_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO item_history (
        item_id,
        action,
        changed_by_user_id,
        changed_by_username,
        changed_by_role,
        old_data,
        new_data
    )
    VALUES (
        NEW.id,
        'INSERT',
        NULLIF(current_setting('app.user_id', true), '')::BIGINT,
        NULLIF(current_setting('app.username', true), ''),
        NULLIF(current_setting('app.role', true), ''),
        NULL,
        to_jsonb(NEW)
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Эта функция логирует обновление товара.
-- OLD содержит старую версию строки.
-- NEW содержит новую версию строки.
CREATE OR REPLACE FUNCTION log_item_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO item_history (
        item_id,
        action,
        changed_by_user_id,
        changed_by_username,
        changed_by_role,
        old_data,
        new_data
    )
    VALUES (
        NEW.id,
        'UPDATE',
        NULLIF(current_setting('app.user_id', true), '')::BIGINT,
        NULLIF(current_setting('app.username', true), ''),
        NULLIF(current_setting('app.role', true), ''),
        to_jsonb(OLD),
        to_jsonb(NEW)
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Эта функция логирует удаление товара.
-- NEW здесь нет, потому что строка удаляется.
-- OLD содержит удаляемую строку товара.
CREATE OR REPLACE FUNCTION log_item_delete()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO item_history (
        item_id,
        action,
        changed_by_user_id,
        changed_by_username,
        changed_by_role,
        old_data,
        new_data
    )
    VALUES (
        OLD.id,
        'DELETE',
        NULLIF(current_setting('app.user_id', true), '')::BIGINT,
        NULLIF(current_setting('app.username', true), ''),
        NULLIF(current_setting('app.role', true), ''),
        to_jsonb(OLD),
        NULL
    );

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;


-- Триггер истории для INSERT.
CREATE TRIGGER items_insert_history_trigger
AFTER INSERT ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_insert();


-- Триггер истории для UPDATE.
CREATE TRIGGER items_update_history_trigger
AFTER UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_update();


-- Триггер истории для DELETE.
CREATE TRIGGER items_delete_history_trigger
AFTER DELETE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_delete();