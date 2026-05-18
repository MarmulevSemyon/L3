DROP TRIGGER IF EXISTS items_delete_history_trigger ON items;
DROP TRIGGER IF EXISTS items_update_history_trigger ON items;
DROP TRIGGER IF EXISTS items_insert_history_trigger ON items;

DROP FUNCTION IF EXISTS log_item_delete();
DROP FUNCTION IF EXISTS log_item_update();
DROP FUNCTION IF EXISTS log_item_insert();

DROP TRIGGER IF EXISTS items_set_updated_at_trigger ON items;
DROP FUNCTION IF EXISTS set_items_updated_at(); 