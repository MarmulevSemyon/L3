CREATE OR REPLACE FUNCTION comments_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('simple', coalesce(NEW.author, '') || ' ' || coalesce(NEW.body, ''));
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER comments_search_vector_trigger
BEFORE INSERT OR UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION comments_search_vector_update();