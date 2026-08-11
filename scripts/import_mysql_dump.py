#!/usr/bin/env python3
import argparse
import configparser
import re
import sys
import time
from pathlib import Path

import pymysql


DATABASE_NAME_PATTERN = re.compile(r"^[A-Za-z0-9_]+$")


def read_database_config(path):
    parser = configparser.ConfigParser()
    if not parser.read(path, encoding="utf-8"):
        raise RuntimeError(f"Cannot read config file: {path}")

    section = parser["default"]
    return {
        "host": section.get("db.host", "127.0.0.1"),
        "port": section.getint("db.port", 3306),
        "user": section.get("db.user", "root"),
        "password": section.get("db.password", ""),
        "charset": "utf8mb4",
        "connect_timeout": 10,
        "read_timeout": 600,
        "write_timeout": 600,
    }


def iter_sql_statements(path):
    buffer = bytearray()
    quote = None

    with path.open("rb") as handle:
        for line in handle:
            if quote is None:
                stripped = line.lstrip()
                if stripped.startswith(b"-- ") or stripped.startswith(b"#"):
                    continue

            index = 0
            while index < len(line):
                char = line[index]
                buffer.append(char)

                if quote is not None:
                    if char == ord("\\") and index + 1 < len(line):
                        index += 1
                        buffer.append(line[index])
                    elif char == quote:
                        if index + 1 < len(line) and line[index + 1] == quote:
                            index += 1
                            buffer.append(line[index])
                        else:
                            quote = None
                elif char in (ord("'"), ord('"'), ord("`")):
                    quote = char
                elif char == ord(";"):
                    statement = bytes(buffer[:-1]).strip()
                    buffer.clear()
                    if statement:
                        yield statement.decode("utf-8")

                index += 1

    remaining = bytes(buffer).strip()
    if remaining:
        yield remaining.decode("utf-8")


def create_database(connection_options, database):
    connection = pymysql.connect(**connection_options, autocommit=True)
    try:
        with connection.cursor() as cursor:
            cursor.execute(
                "SELECT 1 FROM information_schema.schemata WHERE schema_name = %s",
                (database,),
            )
            if cursor.fetchone():
                raise RuntimeError(
                    f"Database {database!r} already exists; refusing to overwrite it"
                )
            cursor.execute(
                f"CREATE DATABASE `{database}` "
                "CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
            )
    finally:
        connection.close()


def import_dump(connection_options, database, source):
    connection = pymysql.connect(
        **connection_options,
        database=database,
        autocommit=True,
    )
    statement_number = 0
    started_at = time.monotonic()

    try:
        with connection.cursor() as cursor:
            for statement_number, statement in enumerate(
                iter_sql_statements(source), start=1
            ):
                cursor.execute(statement)
                if statement_number % 250 == 0:
                    elapsed = time.monotonic() - started_at
                    print(
                        f"Executed {statement_number} statements "
                        f"in {elapsed:.1f}s",
                        flush=True,
                    )
        connection.commit()
    except Exception as error:
        connection.rollback()
        preview = statement[:200].replace("\n", " ") if statement_number else ""
        raise RuntimeError(
            f"Import failed at statement {statement_number}: {preview}"
        ) from error
    finally:
        connection.close()

    return statement_number, time.monotonic() - started_at


def validate_import(connection_options, database):
    connection = pymysql.connect(
        **connection_options,
        database=database,
        autocommit=True,
    )
    try:
        with connection.cursor() as cursor:
            cursor.execute("SHOW TABLES")
            tables = [row[0] for row in cursor.fetchall()]
            counts = []
            for table in tables:
                cursor.execute(f"SELECT COUNT(*) FROM `{table}`")
                counts.append((table, cursor.fetchone()[0]))
            return counts
    finally:
        connection.close()


def main():
    parser = argparse.ArgumentParser(
        description="Import a phpMyAdmin MySQL dump into a new database"
    )
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--database", required=True)
    parser.add_argument("--config", type=Path, default=Path("conf/app.ini"))
    args = parser.parse_args()

    if not args.source.is_file():
        parser.error(f"SQL dump does not exist: {args.source}")
    if not DATABASE_NAME_PATTERN.fullmatch(args.database):
        parser.error("Database name may contain only letters, numbers, and underscores")

    connection_options = read_database_config(args.config)
    print(
        f"Creating {args.database} on "
        f"{connection_options['host']}:{connection_options['port']}",
        flush=True,
    )
    create_database(connection_options, args.database)

    print(f"Importing {args.source} ({args.source.stat().st_size} bytes)", flush=True)
    statement_count, elapsed = import_dump(
        connection_options, args.database, args.source
    )
    print(
        f"Import completed: {statement_count} statements in {elapsed:.1f}s",
        flush=True,
    )

    for table, row_count in validate_import(connection_options, args.database):
        print(f"{table}: {row_count} rows")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"ERROR: {error}", file=sys.stderr)
        if error.__cause__ is not None:
            print(f"CAUSE: {error.__cause__}", file=sys.stderr)
        sys.exit(1)
