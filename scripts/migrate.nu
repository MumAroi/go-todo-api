let DATABASE_URL = "postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable"

def "migrate up" [] {
    migrate -path migrations -database $DATABASE_URL up
}

def "migrate down" [] {
    migrate -path migrations -database $DATABASE_URL down
}

def "migrate goto" [version: int] {
    migrate -path migrations -database $DATABASE_URL goto $version
}

def "migrate version" [] {
    migrate -path migrations -database $DATABASE_URL version
}

def "migrate create" [name: string] {
    ^migrate create -ext sql -dir migrations -seq $name
}

def "migrate force" [version: int] {
    migrate -path migrations -database $DATABASE_URL force $version
}

print "Migration commands loaded:"
print "  migrate up           - Apply all migrations"
print "  migrate down         - Rollback all migrations"
print "  migrate goto N       - Go to specific version"
print "  migrate version      - Show current version"
print "  migrate create NAME  - Create new migration files"
print "  migrate force N      - Force set version, then run migrate up"
