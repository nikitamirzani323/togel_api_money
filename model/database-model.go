package model

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/nikitamirzani323/togel_api_money/config"
	"github.com/nikitamirzani323/togel_api_money/db"
	"github.com/nikitamirzani323/togel_api_money/helpers"
)

func Get_counter(field_column string) int {
	con := db.CreateCon()
	ctx := context.Background()
	idrecord_counter := 0
	sqlcounter := `SELECT 
					counter 
					FROM ` + config.DB_tbl_counter + ` 
					WHERE nmcounter = ? 
				`
	var counter int = 0
	row := con.QueryRowContext(ctx, sqlcounter, field_column)
	switch e := row.Scan(&counter); e {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
	case nil:
		// log.Println(counter)
	default:
		log.Panic(e)
	}
	if counter > 0 {
		idrecord_counter = int(counter) + 1
		stmt, e := con.PrepareContext(ctx, "UPDATE "+config.DB_tbl_counter+" SET counter=? WHERE nmcounter=? ")
		helpers.ErrorCheck(e)
		res, e := stmt.ExecContext(ctx, idrecord_counter, field_column)
		helpers.ErrorCheck(e)
		a, e := res.RowsAffected()
		helpers.ErrorCheck(e)
		if a > 0 {
			// log.Println("UPDATE COUNTER")
		} else {
			log.Panic(e)
		}
	} else {
		stmt, e := con.PrepareContext(ctx, "insert into "+config.DB_tbl_counter+" (nmcounter, counter) values (?, ?)")
		helpers.ErrorCheck(e)
		res, e := stmt.ExecContext(ctx, field_column, 1)
		helpers.ErrorCheck(e)
		id, e := res.RowsAffected()
		helpers.ErrorCheck(e)
		if id > 0 {
			idrecord_counter = 1
		} else {
			log.Panic(e)
		}
		// log.Println("COUNTER Insert id", id)
		// log.Println("NEW")
	}
	return idrecord_counter
}
func Get_mappingdatabase(company string) (string, string, string) {
	tbl_trx_keluarantogel := "db_tot_" + strings.ToLower(company) + ".tbl_trx_keluarantogel"
	tbl_trx_keluarantogel_detail := "db_tot_" + strings.ToLower(company) + ".tbl_trx_keluarantogel_detail"
	view_client := "db_tot_" + strings.ToLower(company) + ".client_view_invoice_" + strings.ToUpper(company)

	return tbl_trx_keluarantogel, tbl_trx_keluarantogel_detail, view_client
}
