-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_created_modified_column_date()
    RETURNS TRIGGER AS
$body$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.created_at = OLD.created_at;
        NEW.updated_at = (now() at time zone 'utc');
    ELSIF TG_OP = 'INSERT' THEN
        NEW.updated_at = (now() at time zone 'utc');
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_created_column_date()
    RETURNS TRIGGER AS
$body$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.created_at = OLD.created_at;
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION update_created_column_date;
DROP FUNCTION update_created_modified_column_date;
