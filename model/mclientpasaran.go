package model

import (
	"context"
	"database/sql"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nikitamirzani323/togel_api_money/config"
	"github.com/nikitamirzani323/togel_api_money/db"
	"github.com/nikitamirzani323/togel_api_money/entities"
	"github.com/nikitamirzani323/togel_api_money/helpers"
	"github.com/nleeper/goment"
)

var mutex sync.RWMutex

func Fetch_Setting() (helpers.Response, error) {
	var obj entities.Model_setting
	var arraobj []entities.Model_setting
	var res helpers.Response

	render_page := time.Now()
	ctx := context.Background()
	con := db.CreateCon()
	var startmaintenance string = ""
	var endmaintenance string = ""
	sql_select := `SELECT 
		startmaintenance, endmaintenance  
		FROM ` + config.DB_tbl_mst_setting + `  
		WHERE idversion  = '1' 
	`
	row := con.QueryRowContext(ctx, sql_select)
	switch e := row.Scan(&startmaintenance, &endmaintenance); e {
	case sql.ErrNoRows:
	case nil:

	default:
		panic(e)
	}

	obj.StartMaintenance = startmaintenance
	obj.EndMaintenance = endmaintenance
	arraobj = append(arraobj, obj)

	res.Status = fiber.StatusOK
	res.Record = arraobj
	res.Totalrecord = len(arraobj)
	res.Time = time.Since(render_page).String()

	return res, nil
}
func Get_Domain() (helpers.Response, error) {
	var obj entities.Model_domain
	var arraobj []entities.Model_domain
	var res helpers.Response
	msg := "Data Not Found"
	render_page := time.Now()
	ctx := context.Background()
	con := db.CreateCon()
	sql_select := `SELECT 
		nmdomain    
		FROM ` + config.DB_tbl_mst_domain + `  
		WHERE tipedomain = 'FRONTEND'
		AND statusdomain ='RUNNING'  
	`
	rowdomain, err := con.QueryContext(ctx, sql_select)
	defer rowdomain.Close()
	helpers.ErrorCheck(err)

	for rowdomain.Next() {
		var nmdomain_db string
		err = rowdomain.Scan(&nmdomain_db)
		if err != nil {
			return res, err
		}
		obj.Domain = nmdomain_db
		arraobj = append(arraobj, obj)
		msg = "Success"
	}

	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = len(arraobj)
	res.Record = arraobj
	res.Time = time.Since(render_page).String()

	return res, nil
}
func FetchAll_MclientPasaran(client_company string) (helpers.Response, error) {
	var obj entities.Model_mclientpasaran
	var arraobj []entities.Model_mclientpasaran
	var res helpers.Response
	var myDays = []string{"minggu", "senin", "selasa", "rabu", "kamis", "jumat", "sabtu"}
	statuspasaran := "OFFLINE"
	msg := "Error"
	render_page := time.Now()
	ctx := context.Background()
	con := db.CreateCon()
	flag := false
	tglnow, _ := goment.New()
	daynow := tglnow.Format("d")
	intVar, _ := strconv.ParseInt(daynow, 0, 8)
	daynowhari := myDays[intVar]
	pasaranhariini := "OFFLINE"
	tbl_trx_keluaran, _, _ := Get_mappingdatabase(client_company)

	sqlpasaran := `SELECT 
		idcomppasaran, idpasarantogel, nmpasarantogel, jamtutup, jamjadwal, jamopen ,pasarandiundi, pasaranurl    
		FROM ` + config.DB_VIEW_CLIENT_VIEW_PASARAN + `  
		WHERE statuspasaranactive = 'Y' 
		AND idcompany = ?
	`

	rowspasaran, err := con.QueryContext(ctx, sqlpasaran, client_company)
	defer rowspasaran.Close()
	helpers.ErrorCheck(err)

	for rowspasaran.Next() {
		pasaranhariini = "OFFLINE"
		statuspasaran = "ONLINE"
		var (
			idcomppasaran                                                                           int
			idpasarantogel, nmpasarantogel, jamtutup, jamjadwal, jamopen, pasarandiundi, pasaranurl string
			tglkeluaran, periodekerluaran, haripasaran                                              string
		)

		err = rowspasaran.Scan(&idcomppasaran, &idpasarantogel, &nmpasarantogel, &jamtutup, &jamjadwal, &jamopen, &pasarandiundi, &pasaranurl)
		if err != nil {
			return res, err
		}

		sqlkeluaran := `
			SELECT 
			datekeluaran, keluaranperiode
			FROM ` + tbl_trx_keluaran + `  
			WHERE idcomppasaran = ?
			ORDER BY datekeluaran DESC
			LIMIT 1
		`
		row := con.QueryRowContext(ctx, sqlkeluaran, idcomppasaran)
		switch err := row.Scan(&tglkeluaran, &periodekerluaran); err {
		case sql.ErrNoRows:
			flag = false
		case nil:
			flag = true
		default:
			flag = false
		}
		if flag {
			jamtutupdoc, _ := goment.New(tglkeluaran)
			sqlpasaranonline := `
				SELECT
					haripasaran
				FROM ` + config.DB_tbl_mst_company_game_pasaran_offline + ` 
				WHERE idcomppasaran = ?
				AND idcompany = ? 
				AND haripasaran = ? 
			`

			errpasaranonline := con.QueryRowContext(ctx, sqlpasaranonline, idcomppasaran, client_company, daynowhari).Scan(&haripasaran)
			jamtutupdoc2 := jamtutupdoc.Format("YYYY-MM-DD") + " " + jamtutup
			taiskrg2 := tglnow.Format("YYYY-MM-DD HH:mm:ss")
			if errpasaranonline != sql.ErrNoRows {
				pasaranhariini = "ONLINE"
				taiskrg := tglnow.Format("YYYY-MM-DD HH:mm:ss")
				jamtutup := tglnow.Format("YYYY-MM-DD") + " " + jamtutup
				jamopen := tglnow.Format("YYYY-MM-DD") + " " + jamopen

				if taiskrg >= jamtutup && taiskrg <= jamopen {
					statuspasaran = "OFFLINE"
				} else {
					statuspasaran = "ONLINE"
				}
				// log.Println(idpasarantogel + " - " + tglnow.Format("YYYY-MM-DD HH:mm:ss") + " - " + jamtutup + " - " + jamopen + " - " + statuspasaran)

			}
			if taiskrg2 > jamtutupdoc2 {
				statuspasaran = "OFFLINE"
			}
			tempcode := periodekerluaran + "-" + idpasarantogel
			log.Printf("tai skrg %s > jamtutp %s", taiskrg2, jamtutupdoc2)
			log.Printf("%s - %s - %s", nmpasarantogel, tempcode, statuspasaran)
		}

		if periodekerluaran != "" {
			obj.PasaranId = idpasarantogel
			obj.PasaranTogel = nmpasarantogel
			obj.PasaranPeriode = "#" + periodekerluaran + "-" + idpasarantogel
			obj.PasaranTglKeluaran = tglkeluaran
			obj.Pasaranmarketclose = tglkeluaran + " " + jamtutup
			obj.Pasaranmarketschedule = tglkeluaran + " " + jamjadwal
			obj.Pasaranmarketopen = tglkeluaran + " " + jamopen
			obj.Pasaranjamtutup = jamtutup
			obj.Pasaranjamopen = jamopen
			obj.Pasarannote = pasarandiundi
			obj.Pasaranurl = pasaranurl
			obj.Pasaranhari = pasaranhariini
			obj.PasaranStatus = statuspasaran
			arraobj = append(arraobj, obj)
			msg = "Success"
		}
		periodekerluaran = ""
	}

	if len(arraobj) > 0 {
		res.Status = fiber.StatusOK
		res.Message = msg
		res.Totalrecord = len(arraobj)
		res.Record = arraobj
		res.Time = time.Since(render_page).String()
	} else {
		res.Status = fiber.StatusBadRequest
		res.Message = "Not Found"
		res.Totalrecord = 0
		res.Record = nil
		res.Time = time.Since(render_page).String()
	}

	return res, nil
}
func FetchAll_MclientPasaranResult(client_company, pasaran_code string) (helpers.Response, error) {
	var obj entities.Model_mclientpasaranResult
	var arraobj []entities.Model_mclientpasaranResult
	var res helpers.Response
	msg := "Error"
	render_page := time.Now()
	con := db.CreateCon()
	ctx := context.Background()
	tbl_trx_keluaran, _, _ := Get_mappingdatabase(client_company)
	sqlresult := `SELECT 
		A.keluaranperiode, A.datekeluaran, A.keluarantogel, B.idpasarantogel 
		FROM ` + tbl_trx_keluaran + ` as A 
		JOIN ` + config.DB_tbl_mst_company_game_pasaran + ` as B ON B.idcomppasaran = A.idcomppasaran
		WHERE B.idcompany = ? 
		AND B.idpasarantogel = ?
		AND A.keluarantogel != '' 
		ORDER BY A.datekeluaran DESC LIMIT 93
	`

	rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
	defer rowresult.Close()
	helpers.ErrorCheck(err)
	norecord := 0
	for rowresult.Next() {
		norecord = norecord + 1
		var (
			keluaranperiode                             string
			datekeluaran, keluarantogel, idpasarantogel string
		)

		err = rowresult.Scan(&keluaranperiode, &datekeluaran, &keluarantogel, &idpasarantogel)
		helpers.ErrorCheck(err)

		obj.No = norecord
		obj.Date = datekeluaran
		obj.Periode = idpasarantogel + "-" + keluaranperiode
		obj.Result = keluarantogel
		arraobj = append(arraobj, obj)
		msg = "Success"
	}
	if len(arraobj) > 0 {
		res.Status = fiber.StatusOK
		res.Message = msg
		res.Totalrecord = len(arraobj)
		res.Record = arraobj
		res.Time = time.Since(render_page).String()
	} else {
		res.Status = fiber.StatusBadRequest
		res.Message = "Not Found"
		res.Totalrecord = 0
		res.Record = nil
		res.Time = time.Since(render_page).String()
	}

	return res, nil
}
func FetchAll_MclientPasaranResultAll(client_company string) (helpers.Response, error) {
	var obj entities.Model_mclientpasaranResultAll
	var arraobj []entities.Model_mclientpasaranResultAll
	var res helpers.Response
	msg := "Error"
	render_page := time.Now()
	con := db.CreateCon()
	ctx := context.Background()
	flag := false
	tbl_trx_keluaran, _, _ := Get_mappingdatabase(client_company)
	sql_listpasarancompany := `SELECT 
		idcomppasaran, idpasarantogel, nmpasarantogel
		FROM ` + config.DB_VIEW_CLIENT_VIEW_PASARAN + ` 
		WHERE idcompany = ? 
		AND statuspasaranactive = 'Y' 
		ORDER BY nmpasarantogel ASC 
	`

	rowresult, err := con.QueryContext(ctx, sql_listpasarancompany, client_company)
	defer rowresult.Close()
	helpers.ErrorCheck(err)
	norecord := 0
	for rowresult.Next() {
		norecord = norecord + 1
		var (
			idcomppasaran_db                                      int
			idpasarantogel_db, nmpasarantogel_db                  string
			tglkeluaran_db, periodekerluaran_db, keluarantogel_db string
		)

		err = rowresult.Scan(&idcomppasaran_db, &idpasarantogel_db, &nmpasarantogel_db)
		helpers.ErrorCheck(err)

		sqlkeluaran := `
			SELECT 
			datekeluaran, keluaranperiode, keluarantogel
			FROM ` + tbl_trx_keluaran + `  
			WHERE idcomppasaran = ?
			AND keluarantogel != "" 
			ORDER BY datekeluaran DESC
			LIMIT 1
		`
		row := con.QueryRowContext(ctx, sqlkeluaran, idcomppasaran_db)
		switch err := row.Scan(&tglkeluaran_db, &periodekerluaran_db, &keluarantogel_db); err {
		case sql.ErrNoRows:
			flag = false
		case nil:
			flag = true
		default:
			flag = false
		}

		if flag {
			obj.No = norecord
			obj.Date = tglkeluaran_db
			obj.Pasaran = nmpasarantogel_db
			obj.Pasaran_code = idpasarantogel_db
			obj.Periode = idpasarantogel_db + "-" + periodekerluaran_db
			obj.Result = keluarantogel_db
			arraobj = append(arraobj, obj)
			msg = "Success"
		}
	}
	if len(arraobj) > 0 {
		res.Status = fiber.StatusOK
		res.Message = msg
		res.Totalrecord = len(arraobj)
		res.Record = arraobj
		res.Time = time.Since(render_page).String()
	} else {
		res.Status = fiber.StatusBadRequest
		res.Message = "Not Found"
		res.Totalrecord = 0
		res.Record = nil
		res.Time = time.Since(render_page).String()
	}

	return res, nil
}
func CheckPasaran(client_company, pasaran_code string) (helpers.Response, error) {
	var obj entities.Model_mclientpasaranCheckPasaran
	var arraobj []entities.Model_mclientpasaranCheckPasaran
	var res helpers.Response
	var myDays = []string{"minggu", "senin", "selasa", "rabu", "kamis", "jumat", "sabtu"}
	statuspasaran := "ONLINE"
	render_page := time.Now()
	msg := "Error"
	con := db.CreateCon()
	ctx := context.Background()

	tglnow, _ := goment.New()
	daynow := tglnow.Format("d")
	intVar, _ := strconv.ParseInt(daynow, 0, 8)
	daynowhari := myDays[intVar]

	tbl_trx_keluaran, _, _ := Get_mappingdatabase(client_company)

	sqlpasaran := `SELECT 
		idcomppasaran, nmpasarantogel, 
		jamtutup, jamopen  
		FROM ` + config.DB_VIEW_CLIENT_VIEW_PASARAN + `  
		WHERE idcompany = ? 
		AND idpasarantogel = ? 
	`

	rowpasaran, err := con.QueryContext(ctx, sqlpasaran, client_company, pasaran_code)
	defer rowpasaran.Close()
	helpers.ErrorCheck(err)
	for rowpasaran.Next() {
		var (
			idcomppasaran, nmpasarantogel, jamtutup, jamopen string
			idtrxkeluaran, keluaranperiode, haripasaran      string
		)

		err = rowpasaran.Scan(&idcomppasaran, &nmpasarantogel, &jamtutup, &jamopen)
		helpers.ErrorCheck(err)

		sqlkeluaran := `
			SELECT 
			idtrxkeluaran, keluaranperiode
			FROM ` + tbl_trx_keluaran + `  
			WHERE idcompany = ?
			AND idcomppasaran = ?
			AND keluarantogel = ''
			LIMIT 1
		`
		err := con.QueryRowContext(ctx, sqlkeluaran, client_company, idcomppasaran).Scan(&idtrxkeluaran, &keluaranperiode)
		helpers.ErrorCheck(err)

		sqlpasaranonline := `
			SELECT
				haripasaran
			FROM ` + config.DB_tbl_mst_company_game_pasaran_offline + ` 
			WHERE idcomppasaran = ?
			AND idcompany = ? 
			AND haripasaran = ? 
		`

		errpasaranonline := con.QueryRowContext(ctx, sqlpasaranonline, idcomppasaran, client_company, daynowhari).Scan(&haripasaran)

		if errpasaranonline != sql.ErrNoRows {
			taiskrg := tglnow.Format("YYYY-MM-DD HH:mm:ss")
			jamtutup := tglnow.Format("YYYY-MM-DD") + " " + jamtutup
			jamopen := tglnow.Format("YYYY-MM-DD") + " " + jamopen

			// intNow, _ := strconv.Atoi(nowconvert)
			// intTutup, _ := strconv.Atoi(tutupconvert)
			// intOpen, _ := strconv.Atoi(openconvert)

			// if intNow > intTutup && intNow < intOpen {
			// 	statuspasaran = "OFFLINE"
			// }

			if taiskrg >= jamtutup && taiskrg <= jamopen {
				statuspasaran = "OFFLINE"
			} else {
				statuspasaran = "ONLINE"
			}
			// log.Println(idpasarantogel + " - " + tglnow.Format("YYYY-MM-DD HH:mm:ss") + " - " + jamtutup + " - " + jamopen + " - " + statuspasaran)
			// log.Println(nowconvert + " - " + tutupconvert + " - " + openconvert + " - " + statuspasaran)
		}

		obj.PasaranIdtansaction = idtrxkeluaran
		obj.PasaranName = nmpasarantogel
		obj.PasaranPeriode = keluaranperiode
		obj.PasaranIdcomp = idcomppasaran
		obj.PasaranStatus = statuspasaran
		arraobj = append(arraobj, obj)
		msg = "Success"
	}
	if len(arraobj) > 0 {
		res.Status = fiber.StatusOK
		res.Message = msg
		res.Totalrecord = len(arraobj)
		res.Record = arraobj
		res.Time = time.Since(render_page).String()
	} else {
		res.Status = fiber.StatusBadRequest
		res.Message = "Not Found"
		res.Totalrecord = 0
		res.Record = nil
		res.Time = time.Since(render_page).String()
	}

	return res, nil
}
func FetchAll_MinitPasaran(client_company, pasaran_code, permainan string) (helpers.Response, error) {
	var res helpers.Response
	msg := "Error"
	con := db.CreateCon()
	render_page := time.Now()
	ctx := context.Background()

	switch permainan {
	case "4-3-2":
		var obj entities.Model_mpasarantogel432
		var arraobj []entities.Model_mpasarantogel432
		sqlresult := `SELECT 
			1_minbet as min_bet, 
			1_maxbet4d as max4d_bet, 1_maxbet3d as max3d_bet,  1_maxbet3dd as max3dd_bet,
			1_maxbet2d as max2d_bet, 1_maxbet2dd as max2dd_bet, 1_maxbet2dt as max2dt_bet, 
			1_maxbet4d_fullbb as maxbet4d_fullbb_bet, 1_maxbet3d_fullbb as maxbet3d_fullbb_bet, 1_maxbet3dd_fullbb as maxbet3dd_fullbb_bet, 
			1_maxbet2d_fullbb as maxbet2d_fullbb_bet, 1_maxbet2dd_fullbb as maxbet2dd_fullbb_bet, 1_maxbet2dt_fullbb as maxbet2dt_fullbb_bet, 
			1_maxbuy4d as maxbuy4d_bet, 1_maxbuy3d as maxbuy3d_bet, 1_maxbuy3dd as maxbuy3dd_bet, 
			1_maxbuy2d as maxbuy2d_bet, 1_maxbuy2dd as maxbuy2dd_bet, 1_maxbuy2dt as maxbuy2dt_bet, 
			1_disc4d as disc4d_bet, 1_disc3d as disc3d_bet, 1_disc3dd as disc3dd_bet, 
			1_disc2d as disc2d_bet, 1_disc2dd as disc2dd_bet, 1_disc2dt as disc2dt_bet, 
			1_win4d as win4d_bet, 1_win3d as win3d_bet, 1_win3dd as win3dd_bet, 
			1_win2d as win2d_bet, 1_win2dd as win2dd_bet, 1_win2dt as win2dt_bet, 
			1_win4dnodisc as win4dnodiskon_bet, 1_win3dnodisc as win3dnodiskon_bet, 1_win3ddnodisc as win3ddnodiskon_bet, 
			1_win2dnodisc as win2dnodiskon_bet, 1_win2ddnodisc as win2ddnodiskon_bet, 1_win2dtnodisc as win2dtnodiskon_bet, 
			1_win4dbb_kena as win4dbb_kena_bet, 1_win3dbb_kena as win3dbb_kena_bet, 1_win3ddbb_kena as win3ddbb_kena_bet, 
			1_win2dbb_kena as win2dbb_kena_bet, 1_win2ddbb_kena as win2ddbb_kena_bet, 1_win2dtbb_kena as win2dtbb_kena_bet,
			1_win4dbb as win4dbb_bet, 1_win3dbb as win3dbb_bet, 1_win3ddbb as win3ddbb_bet, 
			1_win2dbb as win2dbb_bet, 1_win2ddbb as win2ddbb_bet, 1_win2dtbb as win2dtbb_bet,
			1_limittotal4d as limittotal4d_bet, 1_limittotal3d as limittotal3d_bet, 1_limittotal3dd as limittotal3dd_bet, 
			1_limittotal2d as limittotal2d_bet, 1_limittotal2dd as limittotal2dd_bet, 
			1_limittotal2dt as limittotal2dt_bet, 
			1_limittotal4d_fullbb as limittotal4d_fullbb_bet, 1_limittotal3d_fullbb as limittotal3d_fullbb_bet, 1_limittotal3dd_fullbb as limittotal3dd_fullbb_bet, 
			1_limittotal2d_fullbb as limittotal2d_fullbb_bet, 1_limittotal2dd_fullbb as limittotal2dd_fullbb_bet, 1_limittotal2dt_fullbb as limittotal2dt_fullbb_bet, 
			limitline_4d, limitline_3d, limitline_3dd, limitline_2d, limitline_2dd, limitline_2dt, bbfs 
			FROM ` + config.DB_tbl_mst_company_game_pasaran + `  
			WHERE idcompany = ? 
			AND idpasarantogel = ?
		`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()
		helpers.ErrorCheck(err)

		for rowresult.Next() {
			var (
				min_bet, max4d_bet, max3d_bet, max3dd_bet, max2d_bet, max2dd_bet, max2dt_bet                                                                            float32
				maxbet4d_fullbb_bet, maxbet3d_fullbb_bet, maxbet3dd_fullbb_bet, maxbet2d_fullbb_bet, maxbet2dd_fullbb_bet, maxbet2dt_fullbb_bet                         float32
				maxbuy4d_bet, maxbuy3d_bet, maxbuy3dd_bet, maxbuy2d_bet, maxbuy2dd_bet, maxbuy2dt_bet                                                                   float32
				disc4d_bet, disc3d_bet, disc3dd_bet, disc2d_bet, disc2dd_bet, disc2dt_bet                                                                               float32
				win4d_bet, win3d_bet, win3dd_bet, win2d_bet, win2dd_bet, win2dt_bet                                                                                     float32
				win4dnodiskon_bet, win3dnodiskon_bet, win3ddnodiskon_bet, win2dnodiskon_bet, win2ddnodiskon_bet, win2dtnodiskon_bet                                     float32
				win4dbb_kena_bet, win3dbb_kena_bet, win3ddbb_kena_bet, win2dbb_kena_bet, win2ddbb_kena_bet, win2dtbb_kena_bet                                           float32
				win4dbb_bet, win3dbb_bet, win3ddbb_bet, win2dbb_bet, win2ddbb_bet, win2dtbb_bet                                                                         float32
				limittotal4d_bet, limittotal3d_bet, limittotal3dd_bet, limittotal2d_bet, limittotal2dd_bet, limittotal2dt_bet                                           float32
				limittotal4d_fullbb_bet, limittotal3d_fullbb_bet, limittotal3dd_fullbb_bet, limittotal2d_fullbb_bet, limittotal2dd_fullbb_bet, limittotal2dt_fullbb_bet float32
				limitline_4d, limitline_3d, limitline_3dd, limitline_2d, limitline_2dd, limitline_2dt                                                                   uint32
				bbfs                                                                                                                                                    uint8
			)

			err = rowresult.Scan(
				&min_bet, &max4d_bet, &max3d_bet, &max3dd_bet, &max2d_bet, &max2dd_bet, &max2dt_bet,
				&maxbet4d_fullbb_bet, &maxbet3d_fullbb_bet, &maxbet3dd_fullbb_bet, &maxbet2d_fullbb_bet, &maxbet2dd_fullbb_bet, &maxbet2dt_fullbb_bet,
				&maxbuy4d_bet, &maxbuy3d_bet, &maxbuy3dd_bet, &maxbuy2d_bet, &maxbuy2dd_bet, &maxbuy2dt_bet,
				&disc4d_bet, &disc3d_bet, &disc3dd_bet, &disc2d_bet, &disc2dd_bet, &disc2dt_bet,
				&win4d_bet, &win3d_bet, &win3dd_bet, &win2d_bet, &win2dd_bet, &win2dt_bet,
				&win4dnodiskon_bet, &win3dnodiskon_bet, &win3ddnodiskon_bet, &win2dnodiskon_bet, &win2ddnodiskon_bet, &win2dtnodiskon_bet,
				&win4dbb_kena_bet, &win3dbb_kena_bet, &win3ddbb_kena_bet, &win2dbb_kena_bet, &win2ddbb_kena_bet, &win2dtbb_kena_bet,
				&win4dbb_bet, &win3dbb_bet, &win3ddbb_bet, &win2dbb_bet, &win2ddbb_bet, &win2dtbb_bet,
				&limittotal4d_bet, &limittotal3d_bet, &limittotal3dd_bet, &limittotal2d_bet, &limittotal2dd_bet, &limittotal2dt_bet,
				&limittotal4d_fullbb_bet, &limittotal3d_fullbb_bet, &limittotal3dd_fullbb_bet, &limittotal2d_fullbb_bet, &limittotal2dd_fullbb_bet, &limittotal2dt_fullbb_bet,
				&limitline_4d, &limitline_3d, &limitline_3dd, &limitline_2d, &limitline_2dd, &limitline_2dt,
				&bbfs)
			helpers.ErrorCheck(err)
			obj.Min_bet = min_bet
			obj.Max4d_bet = max4d_bet
			obj.Max3d_bet = max3d_bet
			obj.Max3dd_bet = max3dd_bet
			obj.Max2d_bet = max2d_bet
			obj.Max2dd_bet = max2dd_bet
			obj.Max2dt_bet = max2dt_bet
			obj.Max4d_fullbb_bet = maxbet4d_fullbb_bet
			obj.Max3d_fullbb_bet = maxbet3d_fullbb_bet
			obj.Max3dd_fullbb_bet = maxbet3dd_fullbb_bet
			obj.Max2d_fullbb_bet = maxbet2d_fullbb_bet
			obj.Max2dd_fullbb_bet = maxbet2dd_fullbb_bet
			obj.Max2dt_fullbb_bet = maxbet2dt_fullbb_bet
			obj.Max4d_buy = maxbuy4d_bet
			obj.Max3d_buy = maxbuy3d_bet
			obj.Max3dd_buy = maxbuy3dd_bet
			obj.Max2d_buy = maxbuy2d_bet
			obj.Max2dd_buy = maxbuy2dd_bet
			obj.Max2dt_buy = maxbuy2dt_bet
			obj.Disc4d_bet = disc4d_bet
			obj.Disc3d_bet = disc3d_bet
			obj.Disc3dd_bet = disc3dd_bet
			obj.Disc2d_bet = disc2d_bet
			obj.Disc2dd_bet = disc2dd_bet
			obj.Disc2dt_bet = disc2dt_bet
			obj.Win4d_bet = win4d_bet
			obj.Win3d_bet = win3d_bet
			obj.Win3dd_bet = win3dd_bet
			obj.Win2d_bet = win2d_bet
			obj.Win2dd_bet = win2dd_bet
			obj.Win2dt_bet = win2dt_bet
			obj.Win4dnodiskon_bet = win4dnodiskon_bet
			obj.Win3dnodiskon_bet = win3dnodiskon_bet
			obj.Win3ddnodiskon_bet = win3ddnodiskon_bet
			obj.Win2dnodiskon_bet = win2dnodiskon_bet
			obj.Win2ddnodiskon_bet = win2ddnodiskon_bet
			obj.Win2dtnodiskon_bet = win2dtnodiskon_bet
			obj.Win4dbb_kena_bet = win4dbb_kena_bet
			obj.Win3dbb_kena_bet = win3dbb_kena_bet
			obj.Win3ddbb_kena_bet = win3ddbb_kena_bet
			obj.Win2dbb_kena_bet = win2dbb_kena_bet
			obj.Win2ddbb_kena_bet = win2ddbb_kena_bet
			obj.Win2dtbb_kena_bet = win2dtbb_kena_bet
			obj.Win4dbb_bet = win4dbb_bet
			obj.Win3dbb_bet = win3dbb_bet
			obj.Win3ddbb_bet = win3ddbb_bet
			obj.Win2dbb_bet = win2dbb_bet
			obj.Win2ddbb_bet = win2ddbb_bet
			obj.Win2dtbb_bet = win2dtbb_bet
			obj.Limittotal4d_bet = limittotal4d_bet
			obj.Limittotal3d_bet = limittotal3d_bet
			obj.Limittotal3dd_bet = limittotal3dd_bet
			obj.Limittotal2d_bet = limittotal2d_bet
			obj.Limittotal2dd_bet = limittotal2dd_bet
			obj.Limittotal2dt_bet = limittotal2dt_bet
			obj.Limittotal4d_fullbb_bet = limittotal4d_fullbb_bet
			obj.Limittotal3d_fullbb_bet = limittotal3d_fullbb_bet
			obj.Limittotal3dd_fullbb_bet = limittotal3dd_fullbb_bet
			obj.Limittotal2d_fullbb_bet = limittotal2d_fullbb_bet
			obj.Limittotal2dd_fullbb_bet = limittotal2dd_fullbb_bet
			obj.Limittotal2dt_fullbb_bet = limittotal2dt_fullbb_bet
			obj.Limitline_4d = limitline_4d
			obj.Limitline_3d = limitline_3d
			obj.Limitline_3dd = limitline_3dd
			obj.Limitline_2d = limitline_2d
			obj.Limitline_2dd = limitline_2dd
			obj.Limitline_2dt = limitline_2dt
			obj.Bbfs = bbfs

			arraobj = append(arraobj, obj)
			msg = "Success"
		}

		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	case "colok":
		var obj entities.Model_mpasarantogelColok
		var arraobj []entities.Model_mpasarantogelColok
		sqlresult := `SELECT
			2_minbet as min_bet_colokbebas,
			2_maxbet as max_bet_colokbebas,
			2_maxbuy as max_buy_colokbebas,
			2_disc as disc_bet_colokbebas,
			2_win as win_bet_colokbebas, 2_limitotal as limittotal_bet_colokbebas,
			3_minbet as min_bet_colokmacau,
			3_maxbet as max_bet_colokmacau,
			3_maxbuy as max_buy_colokmacau,
			3_disc as disc_bet_colokmacau,
			3_win2digit as win_bet_colokmacau,
			3_win3digit as win3_bet_colokmacau,
			3_win4digit as win4_bet_colokmacau, 3_limittotal as limittotal_bet_colokmacau,
			4_minbet as min_bet_coloknaga,
			4_maxbet as max_bet_coloknaga,
			4_maxbuy as max_buy_coloknaga,
			4_disc as disc_bet_coloknaga,
			4_win3digit as win_bet_coloknaga,
			4_win4digit as win4_bet_coloknaga, 4_limittotal as limittotal_bet_coloknaga,
			5_minbet as min_bet_colokjitu,
			5_maxbet as max_bet_colokjitu,
			5_maxbuy as max_buy_colokjitu,
			5_desic as disc_bet_colokjitu,
			5_winas as winas_bet_colokjitu,
			5_winkop as winkop_bet_colokjitu,
			5_winkepala as winkepala_bet_colokjitu,
			5_winekor as winekor_bet_colokjitu, 5_limitotal as limittotal_bet_colokjitu
			FROM ` + config.DB_tbl_mst_company_game_pasaran + ` 
			WHERE idcompany = ?
			AND idpasarantogel = ?
		`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()

		helpers.ErrorCheck(err)
		for rowresult.Next() {
			var (
				min_bet_colokbebas, max_bet_colokbebas, max_buy_colokbebas, disc_bet_colokbebas, win_bet_colokbebas, limittotal_bet_colokbebas                                                                   float32
				min_bet_colokmacau, max_bet_colokmacau, max_buy_colokmacau, disc_bet_colokmacau, win_bet_colokmacau, win3_bet_colokmacau, win4_bet_colokmacau, limittotal_bet_colokmacau                         float32
				min_bet_coloknaga, max_bet_coloknaga, max_buy_coloknaga, disc_bet_coloknaga, win_bet_coloknaga, win4_bet_coloknaga, limittotal_bet_coloknaga                                                     float32
				min_bet_colokjitu, max_bet_colokjitu, max_buy_colokjitu, disc_bet_colokjitu, winas_bet_colokjitu, winkop_bet_colokjitu, winkepala_bet_colokjitu, winekor_bet_colokjitu, limittotal_bet_colokjitu float32
			)

			err = rowresult.Scan(
				&min_bet_colokbebas, &max_bet_colokbebas, &max_buy_colokbebas, &disc_bet_colokbebas, &win_bet_colokbebas, &limittotal_bet_colokbebas,
				&min_bet_colokmacau, &max_bet_colokmacau, &max_buy_colokmacau, &disc_bet_colokmacau, &win_bet_colokmacau, &win3_bet_colokmacau, &win4_bet_colokmacau, &limittotal_bet_colokmacau,
				&min_bet_coloknaga, &max_bet_coloknaga, &max_buy_coloknaga, &disc_bet_coloknaga, &win_bet_coloknaga, &win4_bet_coloknaga, &limittotal_bet_coloknaga,
				&min_bet_colokjitu, &max_bet_colokjitu, &max_buy_colokjitu, &disc_bet_colokjitu, &winas_bet_colokjitu, &winkop_bet_colokjitu,
				&winkepala_bet_colokjitu, &winekor_bet_colokjitu, &limittotal_bet_colokjitu)
			helpers.ErrorCheck(err)
			obj.Min_bet_colokbebas = min_bet_colokbebas
			obj.Max_bet_colokbebas = max_bet_colokbebas
			obj.Max_buy_colokbebas = max_buy_colokbebas
			obj.Disc_bet_colokbebas = disc_bet_colokbebas
			obj.Win_bet_colokbebas = win_bet_colokbebas
			obj.Limittotal_bet_colokbebas = limittotal_bet_colokbebas
			obj.Min_bet_colokmacau = min_bet_colokmacau
			obj.Max_bet_colokmacau = max_bet_colokmacau
			obj.Max_buy_colokmacau = max_buy_colokmacau
			obj.Disc_bet_colokmacau = disc_bet_colokmacau
			obj.Win_bet_colokmacau = win_bet_colokmacau
			obj.Win3_bet_colokmacau = win3_bet_colokmacau
			obj.Win4_bet_colokmacau = win4_bet_colokmacau
			obj.Limittotal_bet_colokmacau = limittotal_bet_colokmacau
			obj.Min_bet_coloknaga = min_bet_coloknaga
			obj.Max_bet_coloknaga = max_bet_coloknaga
			obj.Max_buy_coloknaga = max_buy_coloknaga
			obj.Disc_bet_coloknaga = disc_bet_coloknaga
			obj.Win_bet_coloknaga = win_bet_coloknaga
			obj.Win4_bet_coloknaga = win4_bet_coloknaga
			obj.Limittotal_bet_coloknaga = limittotal_bet_coloknaga
			obj.Min_bet_colokjitu = min_bet_colokjitu
			obj.Max_bet_colokjitu = max_bet_colokjitu
			obj.Max_buy_colokjitu = max_buy_colokjitu
			obj.Disc_bet_colokjitu = disc_bet_colokjitu
			obj.Winas_bet_colokjitu = winas_bet_colokjitu
			obj.Winkop_bet_colokjitu = winkop_bet_colokjitu
			obj.Winkepala_bet_colokjitu = winkepala_bet_colokjitu
			obj.Winekor_bet_colokjitu = winekor_bet_colokjitu
			obj.Limittotal_bet_colokjitu = limittotal_bet_colokjitu

			arraobj = append(arraobj, obj)
			msg = "Success"
		}
		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	case "5050":
		var obj entities.Model_pasarantogel5050
		var arraobj []entities.Model_pasarantogel5050
		sqlresult := `SELECT
			6_minbet as min_bet_5050umum,
			6_maxbet as max_bet_5050umum,
			6_maxbuy as max_buy_5050umum,
			6_keibesar as keibesar_bet_5050umum,
			6_keikecil as keikecil_bet_5050umum, 
			6_keigenap as keigenap_bet_5050umum,
			6_keiganjil as keiganjil_bet_5050umum,
			6_keitengah as keitengah_bet_5050umum,
			6_keitepi as keitepi_bet_5050umum,
			6_discbesar as discbesar_bet_5050umum,
			6_disckecil as disckecil_bet_5050umum,
			6_discgenap as discgenap_bet_5050umum,
			6_discganjil as discganjil_bet_5050umum,
			6_disctengah as disctengah_bet_5050umum,
			6_disctepi as disctepi_bet_5050umum,
			6_limittotal as limittotal_bet_5050umum,
			7_minbet as min_bet_5050special,
			7_maxbet as max_bet_5050special,
			7_maxbuy as max_buy_5050special,
			7_keiasganjil as keiasganjil_bet_5050special,
			7_keiasgenap as keiasgenap_bet_5050special,
			7_keiasbesar as keiasbesar_bet_5050special,
			7_keiaskecil as keiaskecil_bet_5050special, 
			7_keikopganjil as keikopganjil_bet_5050special,
			7_keikopgenap as keikopgenap_bet_5050special,
			7_keikopbesar as keikopbesar_bet_5050special,
			7_keikopkecil as keikopkecil_bet_5050special,
			7_keikepalaganjil as keikepalaganjil_bet_5050special,
			7_keikepalagenap as keikepalagenap_bet_5050special, 
			7_keikepalabesar as keikepalabesar_bet_5050special,
			7_keikepalakecil as keikepalakecil_bet_5050special,
			7_keiekorganjil as keiekorganjil_bet_5050special,
			7_keiekorgenap as keiekorgenap_bet_5050special,
			7_keiekorbesar as keiekorbesar_bet_5050special,
			7_keiekorkecil as keiekorkecil_bet_5050special,
			7_discasganjil as discasganjil_bet_5050special,
			7_discasgenap as discasgenap_bet_5050special, 
			7_discasbesar as discasbesar_bet_5050special,
			7_discaskecil as discaskecil_bet_5050special,
			7_disckopganjil as disckopganjil_bet_5050special,
			7_disckopgenap as disckopgenap_bet_5050special,
			7_disckopbesar as disckopbesar_bet_5050special,
			7_disckopkecil as disckopkecil_bet_5050special,
			7_disckepalaganjil as disckepalaganjil_bet_5050special,
			7_disckepalagenap as disckepalagenap_bet_5050special,
			7_disckepalabesar as disckepalabesar_bet_5050special,
			7_disckepalakecil as disckepalakecil_bet_5050special,
			7_discekorganjil as discekorganjil_bet_5050special,
			7_discekorgenap as discekorgenap_bet_5050special,
			7_discekorbesar as discekorbesar_bet_5050special,
			7_discekorkecil as discekorkecil_bet_5050special,
			7_limittotal as limittotal_bet_5050special,
			8_minbet as min_bet_5050kombinasi,
			8_maxbet as max_bet_5050kombinasi,
			8_maxbuy as max_buy_5050kombinasi,
			8_belakangkeimono as kei_belakangmono_bet_5050kombinasi,
			8_belakangkeistereo as kei_belakangstereo_bet_5050kombinasi,
			8_belakangkeikembang as kei_belakangkembang_bet_5050kombinasi,
			8_belakangkeikempis as kei_belakangkempis_bet_5050kombinasi,
			8_belakangkeikembar as kei_belakangkembar_bet_5050kombinasi,
			8_tengahkeimono as kei_tengahmono_bet_5050kombinasi,
			8_tengahkeistereo as kei_tengahstereo_bet_5050kombinasi,
			8_tengahkeikembang as kei_tengahkembang_bet_5050kombinasi,
			8_tengahkeikempis as kei_tengahkempis_bet_5050kombinasi,
			8_tengahkeikembar as kei_tengahkembar_bet_5050kombinasi,
			8_depankeimono as kei_depanmono_bet_5050kombinasi,
			8_depankeistereo as kei_depanstereo_bet_5050kombinasi,
			8_depankeikembang as kei_depankembang_bet_5050kombinasi,
			8_depankeikempis as kei_depankempis_bet_5050kombinasi,
			8_depankeikembar as kei_depankembar_bet_5050kombinasi,
			8_belakangdiscmono as disc_belakangmono_bet_5050kombinasi,
			8_belakangdiscstereo as disc_belakangstereo_bet_5050kombinasi,
			8_belakangdisckembang as disc_belakangkembang_bet_5050kombinasi,
			8_belakangdisckempis as disc_belakangkempis_bet_5050kombinasi,
			8_belakangdisckembar as disc_belakangkembar_bet_5050kombinasi,
			8_tengahdiscmono as disc_tengahmono_bet_5050kombinasi,
			8_tengahdiscstereo as disc_tengahstereo_bet_5050kombinasi,
			8_tengahdisckembang as disc_tengahkembang_bet_5050kombinasi,
			8_tengahdisckempis as disc_tengahkempis_bet_5050kombinasi,
			8_tengahdisckembar as disc_tengahkembar_bet_5050kombinasi,
			8_depandiscmono as disc_depanmono_bet_5050kombinasi,
			8_depandiscstereo as disc_depanstereo_bet_5050kombinasi,
			8_depandisckembang as disc_depankembang_bet_5050kombinasi,
			8_depandisckempis as disc_depankempis_bet_5050kombinasi,
			8_depandisckembar as disc_depankembar_bet_5050kombinasi,
			8_limittotal as limittotal_bet_5050kombinasi
			FROM ` + config.DB_tbl_mst_company_game_pasaran + ` 
			WHERE idcompany = ?
			AND idpasarantogel = ?
		`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()

		helpers.ErrorCheck(err)
		for rowresult.Next() {
			var (
				min_bet_5050umum, max_bet_5050umum, max_buy_5050umum                                                                                                                                                            float32
				keibesar_bet_5050umum, keikecil_bet_5050umum, keigenap_bet_5050umum, keiganjil_bet_5050umum, keitengah_bet_5050umum, keitepi_bet_5050umum                                                                       float32
				discbesar_bet_5050umum, disckecil_bet_5050umum, discgenap_bet_5050umum, discganjil_bet_5050umum, disctengah_bet_5050umum, disctepi_bet_5050umum, limittotal_bet_5050umum                                        float32
				min_bet_5050special, max_bet_5050special, max_buy_5050special                                                                                                                                                   float32
				keiasganjil_bet_5050special, keiasgenap_bet_5050special, keiasbesar_bet_5050special, keiaskecil_bet_5050special                                                                                                 float32
				keikopganjil_bet_5050special, keikopgenap_bet_5050special, keikopbesar_bet_5050special, keikopkecil_bet_5050special                                                                                             float32
				keikepalaganjil_bet_5050special, keikepalagenap_bet_5050special, keikepalabesar_bet_5050special, keikepalakecil_bet_5050special                                                                                 float32
				keiekorganjil_bet_5050special, keiekorgenap_bet_5050special, keiekorbesar_bet_5050special, keiekorkecil_bet_5050special                                                                                         float32
				discasganjil_bet_5050special, discasgenap_bet_5050special, discasbesar_bet_5050special, discaskecil_bet_5050special                                                                                             float32
				disckopganjil_bet_5050special, disckopgenap_bet_5050special, disckopbesar_bet_5050special, disckopkecil_bet_5050special                                                                                         float32
				disckepalaganjil_bet_5050special, disckepalagenap_bet_5050special, disckepalabesar_bet_5050special, disckepalakecil_bet_5050special                                                                             float32
				discekorganjil_bet_5050special, discekorgenap_bet_5050special, discekorbesar_bet_5050special, discekorkecil_bet_5050special, limittotal_bet_5050special                                                         float32
				min_bet_5050kombinasi, max_bet_5050kombinasi, max_buy_5050kombinasi                                                                                                                                             float32
				kei_belakangmono_bet_5050kombinasi, kei_belakangstereo_bet_5050kombinasi, kei_belakangkembang_bet_5050kombinasi, kei_belakangkempis_bet_5050kombinasi, kei_belakangkembar_bet_5050kombinasi                     float32
				kei_tengahmono_bet_5050kombinasi, kei_tengahstereo_bet_5050kombinasi, kei_tengahkembang_bet_5050kombinasi, kei_tengahkempis_bet_5050kombinasi, kei_tengahkembar_bet_5050kombinasi                               float32
				kei_depanmono_bet_5050kombinasi, kei_depanstereo_bet_5050kombinasi, kei_depankembang_bet_5050kombinasi, kei_depankempis_bet_5050kombinasi, kei_depankembar_bet_5050kombinasi                                    float32
				disc_belakangmono_bet_5050kombinasi, disc_belakangstereo_bet_5050kombinasi, disc_belakangkembang_bet_5050kombinasi, disc_belakangkempis_bet_5050kombinasi, disc_belakangkembar_bet_5050kombinasi                float32
				disc_tengahmono_bet_5050kombinasi, disc_tengahstereo_bet_5050kombinasi, disc_tengahkembang_bet_5050kombinasi, disc_tengahkempis_bet_5050kombinasi, disc_tengahkembar_bet_5050kombinasi                          float32
				disc_depanmono_bet_5050kombinasi, disc_depanstereo_bet_5050kombinasi, disc_depankembang_bet_5050kombinasi, disc_depankempis_bet_5050kombinasi, disc_depankembar_bet_5050kombinasi, limittotal_bet_5050kombinasi float32
			)

			err = rowresult.Scan(
				&min_bet_5050umum, &max_bet_5050umum, &max_buy_5050umum,
				&keibesar_bet_5050umum, &keikecil_bet_5050umum, &keigenap_bet_5050umum, &keiganjil_bet_5050umum, &keitengah_bet_5050umum, &keitepi_bet_5050umum,
				&discbesar_bet_5050umum, &disckecil_bet_5050umum, &discgenap_bet_5050umum, &discganjil_bet_5050umum, &disctengah_bet_5050umum, &disctepi_bet_5050umum, &limittotal_bet_5050umum,
				&min_bet_5050special, &max_bet_5050special, &max_buy_5050special,
				&keiasganjil_bet_5050special, &keiasgenap_bet_5050special, &keiasbesar_bet_5050special, &keiaskecil_bet_5050special,
				&keikopganjil_bet_5050special, &keikopgenap_bet_5050special, &keikopbesar_bet_5050special, &keikopkecil_bet_5050special,
				&keikepalaganjil_bet_5050special, &keikepalagenap_bet_5050special, &keikepalabesar_bet_5050special, &keikepalakecil_bet_5050special,
				&keiekorganjil_bet_5050special, &keiekorgenap_bet_5050special, &keiekorbesar_bet_5050special, &keiekorkecil_bet_5050special,
				&discasganjil_bet_5050special, &discasgenap_bet_5050special, &discasbesar_bet_5050special, &discaskecil_bet_5050special,
				&disckopganjil_bet_5050special, &disckopgenap_bet_5050special, &disckopbesar_bet_5050special, &disckopkecil_bet_5050special,
				&disckepalaganjil_bet_5050special, &disckepalagenap_bet_5050special, &disckepalabesar_bet_5050special, &disckepalakecil_bet_5050special,
				&discekorganjil_bet_5050special, &discekorgenap_bet_5050special, &discekorbesar_bet_5050special, &discekorkecil_bet_5050special, &limittotal_bet_5050special,
				&min_bet_5050kombinasi, &max_bet_5050kombinasi, &max_buy_5050kombinasi,
				&kei_belakangmono_bet_5050kombinasi, &kei_belakangstereo_bet_5050kombinasi, &kei_belakangkembang_bet_5050kombinasi, &kei_belakangkempis_bet_5050kombinasi, &kei_belakangkembar_bet_5050kombinasi,
				&kei_tengahmono_bet_5050kombinasi, &kei_tengahstereo_bet_5050kombinasi, &kei_tengahkembang_bet_5050kombinasi, &kei_tengahkempis_bet_5050kombinasi, &kei_tengahkembar_bet_5050kombinasi,
				&kei_depanmono_bet_5050kombinasi, &kei_depanstereo_bet_5050kombinasi, &kei_depankembang_bet_5050kombinasi, &kei_depankempis_bet_5050kombinasi, &kei_depankembar_bet_5050kombinasi,
				&disc_belakangmono_bet_5050kombinasi, &disc_belakangstereo_bet_5050kombinasi, &disc_belakangkembang_bet_5050kombinasi, &disc_belakangkempis_bet_5050kombinasi, &disc_belakangkembar_bet_5050kombinasi,
				&disc_tengahmono_bet_5050kombinasi, &disc_tengahstereo_bet_5050kombinasi, &disc_tengahkembang_bet_5050kombinasi, &disc_tengahkempis_bet_5050kombinasi, &disc_tengahkembar_bet_5050kombinasi,
				&disc_depanmono_bet_5050kombinasi, &disc_depanstereo_bet_5050kombinasi, &disc_depankembang_bet_5050kombinasi, &disc_depankempis_bet_5050kombinasi, &disc_depankembar_bet_5050kombinasi,
				&limittotal_bet_5050kombinasi)
			helpers.ErrorCheck(err)
			obj.Min_bet_5050umum = min_bet_5050umum
			obj.Max_bet_5050umum = max_bet_5050umum
			obj.Max_buy_5050umum = max_buy_5050umum
			obj.Keibesar_bet_5050umum = keibesar_bet_5050umum
			obj.Keikecil_bet_5050umum = keikecil_bet_5050umum
			obj.Keigenap_bet_5050umum = keigenap_bet_5050umum
			obj.Keiganjil_bet_5050umum = keiganjil_bet_5050umum
			obj.Keitengah_bet_5050umum = keitengah_bet_5050umum
			obj.Keitepi_bet_5050umum = keitepi_bet_5050umum
			obj.Discbesar_bet_5050umum = discbesar_bet_5050umum
			obj.Disckecil_bet_5050umum = disckecil_bet_5050umum
			obj.Discgenap_bet_5050umum = discgenap_bet_5050umum
			obj.Discganjil_bet_5050umum = discganjil_bet_5050umum
			obj.Disctengah_bet_5050umum = disctengah_bet_5050umum
			obj.Disctepi_bet_5050umum = disctepi_bet_5050umum
			obj.Limittotal_bet_5050umum = limittotal_bet_5050umum

			obj.Min_bet_5050special = min_bet_5050special
			obj.Max_bet_5050special = max_bet_5050special
			obj.Max_buy_5050special = max_buy_5050special
			obj.Keiasganjil_bet_5050special = keiasganjil_bet_5050special
			obj.Keiasgenap_bet_5050special = keiasgenap_bet_5050special
			obj.Keiasbesar_bet_5050special = keiasbesar_bet_5050special
			obj.Keiaskecil_bet_5050special = keiaskecil_bet_5050special
			obj.Keikopganjil_bet_5050special = keikopganjil_bet_5050special
			obj.Keikopgenap_bet_5050special = keikopgenap_bet_5050special
			obj.Keikopbesar_bet_5050special = keikopbesar_bet_5050special
			obj.Keikopkecil_bet_5050special = keikopkecil_bet_5050special
			obj.Keikepalaganjil_bet_5050special = keikepalaganjil_bet_5050special
			obj.Keikepalagenap_bet_5050special = keikepalagenap_bet_5050special
			obj.Keikepalabesar_bet_5050special = keikepalabesar_bet_5050special
			obj.Keikepalakecil_bet_5050special = keikepalakecil_bet_5050special
			obj.Keiekorganjil_bet_5050special = keiekorganjil_bet_5050special
			obj.Keiekorgenap_bet_5050special = keiekorgenap_bet_5050special
			obj.Keiekorbesar_bet_5050special = keiekorbesar_bet_5050special
			obj.Keiekorkecil_bet_5050special = keiekorkecil_bet_5050special
			obj.Discasganjil_bet_5050special = discasganjil_bet_5050special
			obj.Discasgenap_bet_5050special = discasgenap_bet_5050special
			obj.Discasbesar_bet_5050special = discasbesar_bet_5050special
			obj.Discaskecil_bet_5050special = discaskecil_bet_5050special
			obj.Disckopganjil_bet_5050special = disckopganjil_bet_5050special
			obj.Disckopgenap_bet_5050special = disckopgenap_bet_5050special
			obj.Disckopbesar_bet_5050special = disckopbesar_bet_5050special
			obj.Disckopkecil_bet_5050special = disckopkecil_bet_5050special
			obj.Disckepalaganjil_bet_5050special = disckepalaganjil_bet_5050special
			obj.Disckepalagenap_bet_5050special = disckepalagenap_bet_5050special
			obj.Disckepalabesar_bet_5050special = disckepalabesar_bet_5050special
			obj.Disckepalakecil_bet_5050special = disckepalakecil_bet_5050special
			obj.Discekorganjil_bet_5050special = discekorganjil_bet_5050special
			obj.Discekorgenap_bet_5050special = discekorgenap_bet_5050special
			obj.Discekorbesar_bet_5050special = discekorbesar_bet_5050special
			obj.Discekorkecil_bet_5050special = discekorkecil_bet_5050special
			obj.Limittotal_bet_5050special = limittotal_bet_5050special
			obj.Min_bet_5050kombinasi = min_bet_5050kombinasi
			obj.Max_bet_5050kombinasi = max_bet_5050kombinasi
			obj.Max_buy_5050kombinasi = max_buy_5050kombinasi
			obj.Kei_belakangmono_bet_5050kombinasi = kei_belakangmono_bet_5050kombinasi
			obj.Kei_belakangstereo_bet_5050kombinasi = kei_belakangstereo_bet_5050kombinasi
			obj.Kei_belakangkembang_bet_5050kombinasi = kei_belakangkembang_bet_5050kombinasi
			obj.Kei_belakangkempis_bet_5050kombinasi = kei_belakangkempis_bet_5050kombinasi
			obj.Kei_belakangkembar_bet_5050kombinasi = kei_belakangkembar_bet_5050kombinasi
			obj.Kei_tengahmono_bet_5050kombinasi = kei_tengahmono_bet_5050kombinasi
			obj.Kei_tengahstereo_bet_5050kombinasi = kei_tengahstereo_bet_5050kombinasi
			obj.Kei_tengahkembang_bet_5050kombinasi = kei_tengahkembang_bet_5050kombinasi
			obj.Kei_tengahkempis_bet_5050kombinasi = kei_tengahkempis_bet_5050kombinasi
			obj.Kei_tengahkembar_bet_5050kombinasi = kei_tengahkembar_bet_5050kombinasi
			obj.Kei_depanmono_bet_5050kombinasi = kei_depanmono_bet_5050kombinasi
			obj.Kei_depanstereo_bet_5050kombinasi = kei_depanstereo_bet_5050kombinasi
			obj.Kei_depankembang_bet_5050kombinasi = kei_depankembang_bet_5050kombinasi
			obj.Kei_depankempis_bet_5050kombinasi = kei_depankempis_bet_5050kombinasi
			obj.Kei_depankembar_bet_5050kombinasi = kei_depankembar_bet_5050kombinasi
			obj.Disc_belakangmono_bet_5050kombinasi = disc_belakangmono_bet_5050kombinasi
			obj.Disc_belakangstereo_bet_5050kombinasi = disc_belakangstereo_bet_5050kombinasi
			obj.Disc_belakangkembang_bet_5050kombinasi = disc_belakangkembang_bet_5050kombinasi
			obj.Disc_belakangkempis_bet_5050kombinasi = disc_belakangkempis_bet_5050kombinasi
			obj.Disc_belakangkembar_bet_5050kombinasi = disc_belakangkembar_bet_5050kombinasi
			obj.Disc_tengahmono_bet_5050kombinasi = disc_tengahmono_bet_5050kombinasi
			obj.Disc_tengahstereo_bet_5050kombinasi = disc_tengahstereo_bet_5050kombinasi
			obj.Disc_tengahkembang_bet_5050kombinasi = disc_tengahkembang_bet_5050kombinasi
			obj.Disc_tengahkempis_bet_5050kombinasi = disc_tengahkempis_bet_5050kombinasi
			obj.Disc_tengahkembar_bet_5050kombinasi = disc_tengahkembar_bet_5050kombinasi
			obj.Disc_depanmono_bet_5050kombinasi = disc_depanmono_bet_5050kombinasi
			obj.Disc_depanstereo_bet_5050kombinasi = disc_depanstereo_bet_5050kombinasi
			obj.Disc_depankembang_bet_5050kombinasi = disc_depankembang_bet_5050kombinasi
			obj.Disc_depankempis_bet_5050kombinasi = disc_depankempis_bet_5050kombinasi
			obj.Disc_depankembar_bet_5050kombinasi = disc_depankembar_bet_5050kombinasi
			obj.Limittotal_bet_5050kombinasi = limittotal_bet_5050kombinasi

			arraobj = append(arraobj, obj)
			msg = "Success"
		}
		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	case "macaukombinasi":
		var obj entities.Model_pasarantogelMacauKombinasi
		var arraobj []entities.Model_pasarantogelMacauKombinasi
		sqlresult := `SELECT 
		9_minbet as min_bet, 
		9_maxbet as max_bet, 
		9_maxbuy as max_buy, 
		9_win as win_bet, 
		9_discount as diskon_bet, 
		9_limittotal as limit_total 
		FROM ` + config.DB_tbl_mst_company_game_pasaran + `  
		WHERE idcompany = ? 
		AND idpasarantogel = ?
	`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()

		helpers.ErrorCheck(err)
		for rowresult.Next() {
			var (
				min_bet, max_bet, max_buy, win_bet, diskon_bet, limit_total float32
			)

			err = rowresult.Scan(&min_bet, &max_bet, &max_buy, &win_bet, &diskon_bet, &limit_total)
			helpers.ErrorCheck(err)
			obj.Min_bet = min_bet
			obj.Max_bet = max_bet
			obj.Max_buy = max_buy
			obj.Win_bet = win_bet
			obj.Diskon_bet = diskon_bet
			obj.Limit_total = limit_total
			arraobj = append(arraobj, obj)
			msg = "Success"
		}

		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	case "dasar":
		var obj entities.Model_mpasarantogelDasar
		var arraobj []entities.Model_mpasarantogelDasar
		sqlresult := `SELECT 
		10_minbet as min_bet, 
		10_maxbet as max_bet, 
		10_maxbuy as max_buy, 
		10_keibesar as kei_besar_bet, 
		10_keikecil as kei_kecil_bet, 
		10_keigenap as kei_genap_bet, 
		10_keiganjil as kei_ganjil_bet, 
		10_discbesar as disc_besar_bet, 
		10_disckecil as disc_kecil_bet, 
		10_discigenap as disc_genap_bet, 
		10_discganjil as disc_ganjil_bet,  
		10_limittotal as limit_total 
		FROM ` + config.DB_tbl_mst_company_game_pasaran + `  
		WHERE idcompany = ? 
		AND idpasarantogel = ?
	`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()

		helpers.ErrorCheck(err)
		for rowresult.Next() {
			var (
				min_bet, max_bet, max_buy, kei_besar_bet, kei_kecil_bet, kei_genap_bet, kei_ganjil_bet float32
				disc_besar_bet, disc_kecil_bet, disc_genap_bet, disc_ganjil_bet, limit_total           float32
			)

			err = rowresult.Scan(
				&min_bet, &max_bet, &max_buy, &kei_besar_bet, &kei_kecil_bet, &kei_genap_bet,
				&kei_ganjil_bet, &disc_besar_bet, &disc_kecil_bet, &disc_genap_bet, &disc_ganjil_bet,
				&limit_total)
			helpers.ErrorCheck(err)
			obj.Min_bet = min_bet
			obj.Max_bet = max_bet
			obj.Max_buy = max_buy
			obj.Kei_besar_bet = kei_besar_bet
			obj.Kei_kecil_bet = kei_kecil_bet
			obj.Kei_genap_bet = kei_genap_bet
			obj.Kei_ganjil_bet = kei_ganjil_bet
			obj.Disc_besar_bet = disc_besar_bet
			obj.Disc_kecil_bet = disc_kecil_bet
			obj.Disc_genap_bet = disc_genap_bet
			obj.Disc_ganjil_bet = disc_ganjil_bet
			obj.Limit_total = limit_total
			arraobj = append(arraobj, obj)
			msg = "Success"
		}

		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	case "shio":
		var obj entities.Model_mpasarantogelShio
		var arraobj []entities.Model_mpasarantogelShio
		sqlresult := `SELECT 
			11_minbet as min_bet, 
			11_maxbet as max_bet, 
			11_maxbuy as max_buy, 
			11_win as win_bet, 
			11_disc as diskon_bet, 
			11_limittotal as limit_total 
			FROM ` + config.DB_tbl_mst_company_game_pasaran + `  
			WHERE idcompany = ? 
			AND idpasarantogel = ?
		`
		rowresult, err := con.QueryContext(ctx, sqlresult, client_company, pasaran_code)
		defer rowresult.Close()

		helpers.ErrorCheck(err)
		for rowresult.Next() {
			var (
				min_bet, max_bet, max_buy, win_bet, diskon_bet, limit_total float32
			)

			err = rowresult.Scan(&min_bet, &max_bet, &max_buy, &win_bet, &diskon_bet, &limit_total)
			helpers.ErrorCheck(err)
			obj.Min_bet = min_bet
			obj.Max_bet = max_bet
			obj.Max_buy = max_buy
			obj.Win_bet = win_bet
			obj.Diskon_bet = diskon_bet
			obj.Limit_total = limit_total
			arraobj = append(arraobj, obj)
			msg = "Success"
		}

		if len(arraobj) > 0 {
			res.Status = fiber.StatusOK
			res.Message = msg
			res.Totalrecord = len(arraobj)
			res.Record = arraobj
			res.Time = time.Since(render_page).String()
		} else {
			res.Status = fiber.StatusBadRequest
			res.Message = "Not Found"
			res.Totalrecord = 0
			res.Record = nil
			res.Time = time.Since(render_page).String()
		}
	}

	return res, nil
}
func Fetch_LimitTransaksiPasaran432(client_username, client_company, tipe_game string, invoice int) (helpers.Response, error) {
	var obj entities.Model_mpasaranLimit
	var res helpers.Response
	con := db.CreateCon()
	ctx := context.Background()
	render_page := time.Now()
	total4d := 0
	total3d := 0
	total3dd := 0
	total2d := 0
	total2dd := 0
	total2dt := 0
	total_colokbebas := 0
	total_coloknaga := 0
	total_colokmacau := 0
	total_colokjitu := 0
	total_5050umum := 0
	total_5050special := 0
	total_5050kombinasi := 0
	total_macaukombinasi := 0
	total_dasar := 0
	total_shio := 0

	total4d_sum := 0
	total3d_sum := 0
	total3dd_sum := 0
	total2d_sum := 0
	total2dd_sum := 0
	total2dt_sum := 0

	total_colokbebas_sum := 0
	total_coloknaga_sum := 0
	total_colokmacau_sum := 0
	total_colokjitu_sum := 0

	total_5050umum_sum := 0
	total_5050special_sum := 0
	total_5050kombinasi_sum := 0

	total_macaukombinasi_sum := 0
	total_dasar_sum := 0
	total_shio_sum := 0

	_, _, view_client := Get_mappingdatabase(client_company)

	sql := `SELECT 
		typegame, bet, diskon, kei  
		FROM ` + view_client + `  
		WHERE idcompany = ? 
		AND username = ?
		AND idtrxkeluaran = ?
	`
	row, err := con.QueryContext(ctx, sql, client_company, client_username, invoice)
	defer row.Close()
	helpers.ErrorCheck(err)
	for row.Next() {
		var (
			typegame          string
			bet_db            int
			diskon_db, kei_db float64
		)
		err = row.Scan(&typegame, &bet_db, &diskon_db, &kei_db)
		helpers.ErrorCheck(err)
		diskonvalue := math.Ceil(float64(bet_db) * diskon_db)
		keivalue := math.Ceil(float64(bet_db) * kei_db)
		bayar := bet_db - int(diskonvalue) - int(keivalue)
		if typegame == "4D" {
			total4d = total4d + 1
			total4d_sum = total4d_sum + int(bayar)
		}
		if typegame == "3D" {
			total3d = total3d + 1
			total3d_sum = total3d_sum + int(bayar)
		}
		if typegame == "3DD" {
			total3dd = total3dd + 1
			total3dd_sum = total3dd_sum + int(bayar)
		}
		if typegame == "2D" {
			total2d = total2d + 1
			total2d_sum = total2d_sum + int(bayar)
		}
		if typegame == "2DD" {
			total2dd = total2dd + 1
			total2dd_sum = total2dd_sum + int(bayar)
		}
		if typegame == "2DT" {
			total2dt = total2dt + 1
			total2dt_sum = total2dt_sum + int(bayar)
		}
		if typegame == "COLOK_BEBAS" {
			total_colokbebas = total_colokbebas + 1
			total_colokbebas_sum = total_colokbebas_sum + int(bayar)
		}
		if typegame == "COLOK_MACAU" {
			total_colokmacau = total_colokmacau + 1
			total_colokmacau_sum = total_colokmacau_sum + int(bayar)
		}
		if typegame == "COLOK_NAGA" {
			total_coloknaga = total_coloknaga + 1
			total_coloknaga_sum = total_coloknaga_sum + int(bayar)
		}
		if typegame == "COLOK_JITU" {
			total_colokjitu = total_colokjitu + 1
			total_colokjitu_sum = total_colokjitu_sum + int(bayar)
		}
		if typegame == "50_50_UMUM" {
			total_5050umum = total_5050umum + 1
			total_5050umum_sum = total_5050umum_sum + int(bayar)
		}
		if typegame == "50_50_SPECIAL" {
			total_5050special = total_5050special + 1
			total_5050special_sum = total_5050special_sum + int(bayar)
		}
		if typegame == "50_50_KOMBINASI" {
			total_5050kombinasi = total_5050kombinasi + 1
			total_5050kombinasi_sum = total_5050kombinasi_sum + int(bayar)
		}
		if typegame == "MACAU_KOMBINASI" {
			total_macaukombinasi = total_macaukombinasi + 1
			total_macaukombinasi_sum = total_macaukombinasi_sum + int(bayar)
		}
		if typegame == "DASAR" {
			total_dasar = total_dasar + 1
			total_dasar_sum = total_dasar_sum + int(bayar)
		}
		if typegame == "SHIO" {
			total_shio = total_shio + 1
			total_shio_sum = total_shio_sum + int(bayar)
		}
	}
	obj.Total_4d = total4d
	obj.Total_3d = total3d
	obj.Total_3dd = total3dd
	obj.Total_2d = total2d
	obj.Total_2dd = total2dd
	obj.Total_2dt = total2dt
	obj.Total_colokbebas = total_colokbebas
	obj.Total_colokmacau = total_colokmacau
	obj.Total_coloknaga = total_coloknaga
	obj.Total_colokjitu = total_colokjitu
	obj.Total_5050umum = total_5050umum
	obj.Total_5050special = total_5050special
	obj.Total_5050kombinasi = total_5050kombinasi
	obj.Total_macaukombinasi = total_macaukombinasi
	obj.Total_dasar = total_dasar
	obj.Total_shio = total_shio
	obj.Total_4d_sum = total4d_sum
	obj.Total_3d_sum = total3d_sum
	obj.Total_3dd_sum = total3dd_sum
	obj.Total_2d_sum = total2d_sum
	obj.Total_2dd_sum = total2dd_sum
	obj.Total_2dt_sum = total2dt_sum
	obj.Total_colokbebas_sum = total_colokbebas_sum
	obj.Total_colokmacau_sum = total_colokmacau_sum
	obj.Total_coloknaga_sum = total_coloknaga_sum
	obj.Total_colokjitu_sum = total_colokjitu_sum
	obj.Total_5050umum_sum = total_5050umum_sum
	obj.Total_5050special_sum = total_5050special_sum
	obj.Total_5050kombinasi_sum = total_5050kombinasi_sum
	obj.Total_macaukombinasi_sum = total_macaukombinasi_sum
	obj.Total_dasar_sum = total_dasar_sum
	obj.Total_shio_sum = total_shio_sum
	res.Status = fiber.StatusOK
	res.Message = "success"
	res.Record = obj
	res.Time = time.Since(render_page).String()
	return res, nil
}
func Fetch_invoicebet(client_username, client_company string, invoice int) (helpers.ResponseCustom, error) {
	var obj entities.Model_mlistinvoicebet
	var arraobj []entities.Model_mlistinvoicebet
	var objgroup entities.Model_mgroupinvoicebetPermainan
	var arraobjgroup []entities.Model_mgroupinvoicebetPermainan
	var res helpers.ResponseCustom
	var totalbayar int = 0

	msg := "Error"
	ctx := context.Background()
	con := db.CreateCon()
	render_page := time.Now()
	_, _, view_client_invoice := Get_mappingdatabase(client_company)

	sql := `SELECT 
		datetimedetail, username, posisitogel, typegame, nomortogel, idpasarantogel, bet, 
		diskon, win, kei, statuskeluarandetail, keluaranperiode
		FROM ` + view_client_invoice + `  
		WHERE idcompany = ? 
		AND username = ?
		AND idtrxkeluaran = ?
		ORDER BY datetimedetail DESC
	`
	row, err := con.QueryContext(ctx, sql, client_company, client_username, invoice)
	defer row.Close()

	helpers.ErrorCheck(err)
	nobet := 0
	for row.Next() {
		nobet = nobet + 1
		var (
			datetimedetail, username, posisitogel, typegame, nomortogel, idpasarantogel string
			bet, diskon, win, kei                                                       float32
			statuskeluarandetail, keluaranperiode                                       string
		)
		err = row.Scan(
			&datetimedetail, &username, &posisitogel, &typegame, &nomortogel,
			&idpasarantogel, &bet, &diskon, &win, &kei, &statuskeluarandetail,
			&keluaranperiode)

		helpers.ErrorCheck(err)
		var diskon2 float32 = diskon * 100
		var diskonbet int = int(bet * diskon)
		var kei2 float32 = kei * 100
		var keibet int = int(bet * kei)
		var menang int = int(bet) * int(win)
		var bayar int = int(bet) - int(diskonbet) - int(keibet)
		totalbayar = int(totalbayar) + int(bayar)

		obj.Tanggal = datetimedetail
		obj.Permainan = typegame
		obj.Tipe = posisitogel
		obj.Periode = idpasarantogel + "-" + keluaranperiode
		obj.Nomor = nomortogel
		obj.Bet = int(bet)
		obj.Diskon = diskon2
		obj.Kei = kei2
		obj.Win = int(win)
		obj.Bayar = bayar
		obj.Menang = menang
		arraobj = append(arraobj, obj)
		msg = "Success"
	}

	sqlgrouppermainan := `SELECT
		typegame
		FROM ` + view_client_invoice + ` 
		WHERE idcompany = ?
		AND username = ?
		AND idtrxkeluaran = ?
		GROUP BY typegame
	`
	rowgrouppermainan, err := con.QueryContext(ctx, sqlgrouppermainan, client_company, client_username, invoice)
	defer rowgrouppermainan.Close()

	helpers.ErrorCheck(err)
	for rowgrouppermainan.Next() {
		var (
			typegame string
		)
		err = rowgrouppermainan.Scan(
			&typegame)

		if err != nil {
			return res, err
		}
		objgroup.Permainan = typegame
		arraobjgroup = append(arraobjgroup, objgroup)
		msg = "Success"
	}

	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = len(arraobj)
	res.Totalbayar = totalbayar
	res.Permainan = arraobjgroup
	res.Record = arraobj
	res.Time = time.Since(render_page).String()
	return res, nil
}
func Fetch_invoicebetbyid(idtrxkeluaran int, client_username, client_company, typegame string) (helpers.Response, error) {
	var obj entities.Model_mlistinvoicebetid
	var arraobj []entities.Model_mlistinvoicebetid
	var res helpers.Response
	flag_3dd := false
	flag_2dd := false
	flag_2dt := false
	msg := "Error"
	ctx := context.Background()
	con := db.CreateCon()
	render_page := time.Now()
	_, _, view_client_invoice := Get_mappingdatabase(client_company)

	sql_select := `SELECT 
		posisitogel, nomortogel, typegame, bet, diskon, kei,winhasil, statuskeluarandetail 
		FROM ` + view_client_invoice + `  
		WHERE idtrxkeluaran = ? 
		AND idcompany = ? 
		AND username = ? 
		ORDER BY nomortogel ASC 
	`
	log.Println(typegame)
	row, err := con.QueryContext(ctx, sql_select, idtrxkeluaran, client_company, client_username)
	defer row.Close()

	helpers.ErrorCheck(err)
	nobet := 0
	for row.Next() {
		var (
			posisitogel_db, nomortogel_db, typegame_db, statuskeluarandetail_db string
			bet_db, diskon_db, kei_db                                           float32
			winhasil_db                                                         int
		)
		err = row.Scan(
			&posisitogel_db, &nomortogel_db, &typegame_db, &bet_db, &diskon_db,
			&kei_db, &winhasil_db, &statuskeluarandetail_db)
		helpers.ErrorCheck(err)

		if typegame == typegame_db {
			nobet = nobet + 1
			var diskon2 float32 = diskon_db * 100
			var diskonbet int = int(bet_db * diskon_db)
			var kei2 float32 = kei_db * 100
			var keibet int = int(bet_db * kei_db)
			var bayar int = int(bet_db) - int(diskonbet) - int(keibet)

			obj.No = nobet
			obj.Status = statuskeluarandetail_db
			obj.Tipe = posisitogel_db
			obj.Permainan = typegame_db
			obj.Nomor = nomortogel_db
			obj.Bet = int(bet_db)
			obj.Diskon = diskon2
			obj.Kei = kei2
			obj.Bayar = bayar
			obj.Win = int(winhasil_db)
			arraobj = append(arraobj, obj)
			msg = "Success"
		}
		if typegame == "2D" {
			flag_2dd = true
			flag_2dt = true
		}
		if typegame == "3D" {
			flag_3dd = true
		}
		if flag_3dd {
			if typegame_db == "3DD" {
				nobet = nobet + 1
				var diskon2 float32 = diskon_db * 100
				var diskonbet int = int(bet_db * diskon_db)
				var kei2 float32 = kei_db * 100
				var keibet int = int(bet_db * kei_db)
				var bayar int = int(bet_db) - int(diskonbet) - int(keibet)

				obj.No = nobet
				obj.Status = statuskeluarandetail_db
				obj.Tipe = posisitogel_db
				obj.Permainan = typegame_db
				obj.Nomor = nomortogel_db
				obj.Bet = int(bet_db)
				obj.Diskon = diskon2
				obj.Kei = kei2
				obj.Bayar = bayar
				obj.Win = int(winhasil_db)
				arraobj = append(arraobj, obj)
				msg = "Success"
			}
		}
		if flag_2dd {
			if typegame_db == "2DD" {
				nobet = nobet + 1
				var diskon2 float32 = diskon_db * 100
				var diskonbet int = int(bet_db * diskon_db)
				var kei2 float32 = kei_db * 100
				var keibet int = int(bet_db * kei_db)
				var bayar int = int(bet_db) - int(diskonbet) - int(keibet)

				obj.No = nobet
				obj.Status = statuskeluarandetail_db
				obj.Tipe = posisitogel_db
				obj.Permainan = typegame_db
				obj.Nomor = nomortogel_db
				obj.Bet = int(bet_db)
				obj.Diskon = diskon2
				obj.Kei = kei2
				obj.Bayar = bayar
				obj.Win = int(winhasil_db)
				arraobj = append(arraobj, obj)
				msg = "Success"
			}
		}
		if flag_2dt {
			if typegame_db == "2DT" {
				nobet = nobet + 1
				var diskon2 float32 = diskon_db * 100
				var diskonbet int = int(bet_db * diskon_db)
				var kei2 float32 = kei_db * 100
				var keibet int = int(bet_db * kei_db)
				var bayar int = int(bet_db) - int(diskonbet) - int(keibet)

				obj.No = nobet
				obj.Status = statuskeluarandetail_db
				obj.Tipe = posisitogel_db
				obj.Permainan = typegame_db
				obj.Nomor = nomortogel_db
				obj.Bet = int(bet_db)
				obj.Diskon = diskon2
				obj.Kei = kei2
				obj.Bayar = bayar
				obj.Win = int(winhasil_db)
				arraobj = append(arraobj, obj)
				msg = "Success"
			}
		}
	}

	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = len(arraobj)
	res.Record = arraobj
	res.Time = time.Since(render_page).String()
	return res, nil
}
func Fetch_invoiceperiode(client_username, client_company, pasaran_code string) (helpers.Response, error) {
	var obj entities.Model_mlistsipperiode
	var arraobj []entities.Model_mlistsipperiode
	var res helpers.Response

	msg := "Error"
	con := db.CreateCon()
	ctx := context.Background()
	render_page := time.Now()
	_, trx_keluarantogel_detail, view_client_invoice := Get_mappingdatabase(client_company)

	sql := `SELECT 
		idtrxkeluaran,datekeluaran,idpasarantogel,keluaranperiode,keluarantogel 
		FROM ` + view_client_invoice + `  
		WHERE idcompany = ? 
		AND username = ? 
		AND idpasarantogel = ? 
		GROUP BY idtrxkeluaran 
		ORDER BY datetimedetail DESC LIMIT 62
	`
	row, err := con.QueryContext(ctx, sql, client_company, client_username, pasaran_code)
	defer row.Close()

	helpers.ErrorCheck(err)
	no := 0
	for row.Next() {
		no = no + 1
		var (
			idtrxkeluaran_DB, datekeluaran_DB, idpasarantogel_DB, keluaranperiode_DB, keluarantogel_DB string
		)
		err = row.Scan(
			&idtrxkeluaran_DB, &datekeluaran_DB, &idpasarantogel_DB, &keluaranperiode_DB,
			&keluarantogel_DB)

		helpers.ErrorCheck(err)
		var idtrxkeluaran string = idtrxkeluaran_DB
		var datekeluaran string = datekeluaran_DB
		var keluarantogel string = keluarantogel_DB
		var periode string = idpasarantogel_DB + "-" + keluaranperiode_DB
		var status string = ""
		var background string = ""
		var totalbet int = 0
		var totalbayar int = 0
		var totalwin int = 0
		var totallose int = 0

		if keluarantogel != "" {
			status = "APPROVED"
		} else {
			status = "RUNNING"
		}
		switch status {
		case "RUNNING":
			background = "background:#FFEB3B;font-size:12px;font-weight:bold;color:black;"
		case "APPROVED":
			background = "background:#1ba573;color:black;font-weight:bold;font-size:12px;"
		}
		if status == "APPROVED" {
			status = "COMPLETED"
		}

		sqldetailbet := `SELECT 
			statuskeluarandetail, typegame, 
			bet, diskon, kei, win 
			FROM ` + trx_keluarantogel_detail + `  
			WHERE idcompany = ? 
			AND username = ? 
			AND idtrxkeluaran = ? 
		`

		rowdetailbet, err := con.QueryContext(ctx, sqldetailbet, client_company, client_username, idtrxkeluaran)
		defer rowdetailbet.Close()

		helpers.ErrorCheck(err)
		for rowdetailbet.Next() {
			totalbet = totalbet + 1

			var (
				statuskeluarandetail_DB, typegame_DB string
				bet_DB, diskon_DB, kei_DB, win_DB    float32
			)
			err = rowdetailbet.Scan(
				&statuskeluarandetail_DB, &typegame_DB,
				&bet_DB, &diskon_DB, &kei_DB, &win_DB)

			helpers.ErrorCheck(err)
			var statuskeluarandetail string = statuskeluarandetail_DB
			var typegame string = typegame_DB
			var bet int = int(bet_DB)
			var diskon float32 = diskon_DB
			var kei float32 = kei_DB
			var win float32 = win_DB
			var bayar int = 0
			var bayarwin int = 0
			var winhasil int = 0
			if typegame == "50_50_UMUM" || typegame == "50_50_SPECIAL" || typegame == "50_50_KOMBINASI" || typegame == "DASAR" || typegame == "COLOK_BEBAS" || typegame == "COLOK_NAGA" || typegame == "COLOK_MACAU" || typegame == "COLOK_JITU" {
				bayar = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				if statuskeluarandetail == "WINNER" {
					bayarwin = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
					winhasil = bayarwin + int(float32(bet)*win)
					totalwin = totalwin + winhasil
				}
			} else {
				bayar = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				if statuskeluarandetail == "WINNER" {
					winhasil = int(float32(bet) * win)
					totalwin = totalwin + winhasil
				}
			}
			totalbayar = totalbayar + bayar
			totallose = totalwin - totalbayar
		}

		obj.Idinvoice = idtrxkeluaran
		obj.Tanggal = datekeluaran
		obj.Periode = periode
		obj.Totalbet = totalbet
		obj.Totalbayar = totalbayar
		obj.Totalwin = totalwin
		obj.Totallose = totallose
		obj.Status = status
		obj.Background = background

		arraobj = append(arraobj, obj)
		msg = "Success"
	}
	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = len(arraobj)
	res.Record = arraobj
	res.Time = time.Since(render_page).String()
	return res, nil
}
func Fetch_invoiceperiodeall(client_username, client_company string) (helpers.Response, error) {
	var obj entities.Model_mlistsipperiodeall
	var arraobj []entities.Model_mlistsipperiodeall
	var res helpers.Response

	msg := "Error"
	con := db.CreateCon()
	ctx := context.Background()
	render_page := time.Now()
	_, trx_keluarantogel_detail, view_client_invoice := Get_mappingdatabase(client_company)
	log.Println(client_username + " " + client_username)
	sql := `SELECT 
		idtrxkeluaran,datekeluaran,idpasarantogel,nmpasarantogel,keluaranperiode,keluarantogel 
		FROM ` + view_client_invoice + `  
		WHERE idcompany = ? 
		AND username = ? 
		GROUP BY idtrxkeluaran 
		ORDER BY datekeluaran DESC LIMIT 31 
	`

	row, err := con.QueryContext(ctx, sql, client_company, client_username)
	defer row.Close()

	helpers.ErrorCheck(err)
	no := 0
	for row.Next() {
		no = no + 1
		var (
			idtrxkeluaran_DB, datekeluaran_DB, idpasarantogel_DB, nmpasarantogel_db, keluaranperiode_DB, keluarantogel_DB string
		)
		err = row.Scan(
			&idtrxkeluaran_DB, &datekeluaran_DB, &idpasarantogel_DB, &nmpasarantogel_db, &keluaranperiode_DB,
			&keluarantogel_DB)

		helpers.ErrorCheck(err)
		var idtrxkeluaran string = idtrxkeluaran_DB
		var datekeluaran string = datekeluaran_DB
		var keluarantogel string = keluarantogel_DB
		var pasarantogel string = nmpasarantogel_db
		var periode string = idpasarantogel_DB + "-" + keluaranperiode_DB
		var status string = ""
		var background string = ""
		var totalbet int = 0
		var totalbayar int = 0
		var totalwin int = 0
		var totallose int = 0

		if keluarantogel != "" {
			status = "APPROVED"
		} else {
			status = "RUNNING"
		}
		switch status {
		case "RUNNING":
			background = "background:#FFEB3B;font-size:12px;font-weight:bold;color:black;"
		case "APPROVED":
			background = "background:#1ba573;color:black;font-weight:bold;font-size:12px;"
		}
		if status == "APPROVED" {
			status = "COMPLETED"
		}

		sqldetailbet := `SELECT 
			statuskeluarandetail, typegame, 
			bet, diskon, kei, win 
			FROM ` + trx_keluarantogel_detail + `  
			WHERE idcompany = ? 
			AND username = ? 
			AND idtrxkeluaran = ? 
		`
		rowdetailbet, err := con.QueryContext(ctx, sqldetailbet, client_company, client_username, idtrxkeluaran)
		defer rowdetailbet.Close()

		helpers.ErrorCheck(err)
		for rowdetailbet.Next() {
			totalbet = totalbet + 1

			var (
				statuskeluarandetail_DB, typegame_DB string
				bet_DB, diskon_DB, kei_DB, win_DB    float32
			)
			err = rowdetailbet.Scan(
				&statuskeluarandetail_DB, &typegame_DB,
				&bet_DB, &diskon_DB, &kei_DB, &win_DB)

			helpers.ErrorCheck(err)
			var statuskeluarandetail string = statuskeluarandetail_DB
			var typegame string = typegame_DB
			var bet int = int(bet_DB)
			var diskon float32 = diskon_DB
			var kei float32 = kei_DB
			var win float32 = win_DB
			var bayar int = 0
			var bayarwin int = 0
			var winhasil int = 0
			if typegame == "50_50_UMUM" || typegame == "50_50_SPECIAL" || typegame == "50_50_KOMBINASI" || typegame == "DASAR" || typegame == "COLOK_BEBAS" || typegame == "COLOK_NAGA" || typegame == "COLOK_MACAU" || typegame == "COLOK_JITU" {
				bayar = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				if statuskeluarandetail == "WINNER" {
					bayarwin = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
					winhasil = bayarwin + int(float32(bet)*win)
					totalwin = totalwin + winhasil
				}
			} else {
				bayar = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				if statuskeluarandetail == "WINNER" {
					winhasil = int(float32(bet) * win)
					totalwin = totalwin + winhasil
				}
			}
			totalbayar = totalbayar + bayar
			totallose = totalwin - totalbayar
		}

		obj.Idinvoice = idtrxkeluaran
		obj.Tanggal = datekeluaran
		obj.Pasaran = pasarantogel
		obj.Periode = periode
		obj.Totalbet = totalbet
		obj.Totalbayar = totalbayar
		obj.Totalwin = totalwin
		obj.Totallose = totallose
		obj.Status = status
		obj.Background = background

		arraobj = append(arraobj, obj)
		msg = "Success"
	}
	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = len(arraobj)
	res.Record = arraobj
	res.Time = time.Since(render_page).String()
	return res, nil
}
func Fetch_invoiceperiodedetail(client_username, client_company, idtrxkeluaran string) (helpers.Response, error) {
	var obj entities.Model_mlistsipperiodedetail
	var res helpers.Response

	msg := "Error"
	con := db.CreateCon()
	ctx := context.Background()
	render_page := time.Now()
	_, trx_keluarantogel_detail, _ := Get_mappingdatabase(client_company)

	sql := `SELECT 
		statuskeluarandetail, typegame, 
		bet, diskon, kei, win 
		FROM ` + trx_keluarantogel_detail + `    
		WHERE idcompany = ? 
		AND username = ?
		AND idtrxkeluaran = ?
	`

	row, err := con.QueryContext(ctx, sql, client_company, client_username, idtrxkeluaran)
	defer row.Close()

	helpers.ErrorCheck(err)
	bayar_4d := 0
	bayar_3d := 0
	bayar_2d := 0
	bayar_colokbebas := 0
	bayar_colokmacau := 0
	bayar_coloknaga := 0
	bayar_colokjitu := 0
	bayar_5050umum := 0
	bayar_5050special := 0
	bayar_5050kombinasi := 0
	bayar_macaukombinasi := 0
	bayar_dasar := 0
	bayar_shio := 0
	total4d_bayar := 0
	total3d_bayar := 0
	total2d_bayar := 0
	totalcolokbebas_bayar := 0
	totalcolokmacau_bayar := 0
	totalcoloknaga_bayar := 0
	totalcolokjitu_bayar := 0
	total5050umum_bayar := 0
	total5050special_bayar := 0
	total5050kombinasi_bayar := 0
	totalmacaukombinasi_bayar := 0
	totaldasar_bayar := 0
	totalshio_bayar := 0
	totalwin_4d := 0
	totalwin_3d := 0
	totalwin_2d := 0
	totalwin_colokbebas := 0
	totalwin_colokmacau := 0
	totalwin_coloknaga := 0
	totalwin_colokjitu := 0
	totalwin_5050umum := 0
	totalwin_5050special := 0
	totalwin_5050kombinasi := 0
	totalwin_macaukombinasi := 0
	totalwin_dasar := 0
	totalwin_shio := 0
	subtotal_bayar := 0
	subtotal_winner := 0
	total_winlose := 0
	for row.Next() {
		var (
			statuskeluarandetail_DB, typegame_DB string
			bet_DB, diskon_DB, kei_DB, win_DB    float32
		)
		err = row.Scan(
			&statuskeluarandetail_DB, &typegame_DB, &bet_DB, &diskon_DB,
			&kei_DB, &win_DB)

		helpers.ErrorCheck(err)
		var statuskeluarandetail string = statuskeluarandetail_DB
		var typegame string = typegame_DB
		var bet int = int(bet_DB)
		var diskon float32 = diskon_DB
		var kei float32 = kei_DB
		var win float32 = win_DB
		var winhasil int = 0
		if typegame == "4D" {
			bayar_4d = bet - int(float32(bet)*diskon)
			total4d_bayar = total4d_bayar + bayar_4d
			if statuskeluarandetail == "WINNER" {
				winhasil = int(float32(bet) * win)
				totalwin_4d = totalwin_4d + winhasil
			}
		}
		if typegame == "3D" || typegame == "3DD" {
			bayar_3d = bet - int(float32(bet)*diskon)
			total3d_bayar = total3d_bayar + bayar_3d
			if statuskeluarandetail == "WINNER" {
				winhasil = int(float32(bet) * win)
				totalwin_3d = totalwin_3d + winhasil
			}
		}
		if typegame == "2D" || typegame == "2DD" || typegame == "2DT" {
			bayar_2d = bet - int(float32(bet)*diskon)
			total2d_bayar = total2d_bayar + bayar_2d
			if statuskeluarandetail == "WINNER" {
				winhasil = int(float32(bet) * win)
				totalwin_2d = totalwin_2d + winhasil
			}
		}
		if typegame == "COLOK_BEBAS" {
			bayar_colokbebas = bet - int(float32(bet)*diskon)
			totalcolokbebas_bayar = totalcolokbebas_bayar + bayar_colokbebas
			if statuskeluarandetail == "WINNER" {
				bayar_colokbebas_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_colokbebas_win + int(float32(bet)*win)
				totalwin_colokbebas = totalwin_colokbebas + winhasil
			}
		}
		if typegame == "COLOK_MACAU" {
			bayar_colokmacau = bet - int(float32(bet)*diskon)
			totalcolokmacau_bayar = totalcolokmacau_bayar + bayar_colokmacau
			if statuskeluarandetail == "WINNER" {
				bayar_colokmacau_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_colokmacau_win + int(float32(bet)*win)
				totalwin_colokmacau = totalwin_colokmacau + winhasil
			}
		}
		if typegame == "COLOK_NAGA" {
			bayar_coloknaga = bet - int(float32(bet)*diskon)
			totalcoloknaga_bayar = totalcoloknaga_bayar + bayar_coloknaga
			if statuskeluarandetail == "WINNER" {
				bayar_coloknaga_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_coloknaga_win + int(float32(bet)*win)
				totalwin_coloknaga = totalwin_coloknaga + winhasil
			}
		}
		if typegame == "COLOK_JITU" {
			bayar_colokjitu = bet - int(float32(bet)*diskon)
			totalcolokjitu_bayar = totalcolokjitu_bayar + bayar_colokjitu
			if statuskeluarandetail == "WINNER" {
				bayar_colokjitu_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_colokjitu_win + int(float32(bet)*win)
				totalwin_colokjitu = totalwin_colokjitu + winhasil
			}
		}
		if typegame == "50_50_UMUM" {
			bayar_5050umum = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			total5050umum_bayar = total5050umum_bayar + bayar_5050umum
			if statuskeluarandetail == "WINNER" {
				bayar_5050umum_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_5050umum_win + int(float32(bet)*win)
				totalwin_5050umum = totalwin_5050umum + winhasil
			}
		}
		if typegame == "50_50_SPECIAL" {
			bayar_5050special = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			total5050special_bayar = total5050special_bayar + bayar_5050special
			if statuskeluarandetail == "WINNER" {
				bayar_5050special_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_5050special_win + int(float32(bet)*win)
				totalwin_5050special = totalwin_5050special + winhasil
			}
		}
		if typegame == "50_50_KOMBINASI" {
			bayar_5050kombinasi = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			total5050kombinasi_bayar = total5050kombinasi_bayar + bayar_5050kombinasi
			if statuskeluarandetail == "WINNER" {
				bayar_5050kombinasi_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_5050kombinasi_win + int(float32(bet)*win)
				totalwin_5050kombinasi = totalwin_5050kombinasi + winhasil
			}
		}
		if typegame == "MACAU_KOMBINASI" {
			bayar_macaukombinasi = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			totalmacaukombinasi_bayar = totalmacaukombinasi_bayar + bayar_macaukombinasi
			if statuskeluarandetail == "WINNER" {
				bayar_macaukombinasi_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_macaukombinasi_win + int(float32(bet)*win)
				totalwin_macaukombinasi = totalwin_macaukombinasi + winhasil
			}
		}
		if typegame == "DASAR" {
			bayar_dasar = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			totaldasar_bayar = totaldasar_bayar + bayar_dasar
			if statuskeluarandetail == "WINNER" {
				bayar_dasar_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_dasar_win + int(float32(bet)*win)
				totalwin_dasar = totalwin_dasar + winhasil
			}
		}
		if typegame == "SHIO" {
			bayar_shio = bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
			totalshio_bayar = totalshio_bayar + bayar_shio
			if statuskeluarandetail == "WINNER" {
				bayar_shio_win := bet - int(float32(bet)*diskon) - int(float32(bet)*kei)
				winhasil = bayar_shio_win + int(float32(bet)*win)
				totalwin_shio = totalwin_shio + winhasil
			}
		}
		msg = "Success"
	}
	subtotal_bayar = total4d_bayar + total3d_bayar + total2d_bayar + totalcolokbebas_bayar + totalcolokmacau_bayar + totalcoloknaga_bayar + totalcolokjitu_bayar + total5050umum_bayar + total5050special_bayar + total5050kombinasi_bayar + totalmacaukombinasi_bayar + totaldasar_bayar + totalshio_bayar
	subtotal_winner = totalwin_4d + totalwin_3d + totalwin_2d + totalwin_colokbebas + totalwin_colokmacau + totalwin_coloknaga + totalwin_colokjitu + totalwin_5050umum + totalwin_5050special + totalwin_5050kombinasi + totalwin_macaukombinasi + totalwin_dasar + totalwin_shio
	total_winlose = subtotal_winner - subtotal_bayar

	obj.Total4d_bayar = total4d_bayar
	obj.Total3d_bayar = total3d_bayar
	obj.Total2d_bayar = total2d_bayar
	obj.Totalcolokbebas_bayar = totalcolokbebas_bayar
	obj.Totalcolokmacau_bayar = totalcolokmacau_bayar
	obj.Totalcoloknaga_bayar = totalcoloknaga_bayar
	obj.Totalcolokjitu_bayar = totalcolokjitu_bayar
	obj.Total5050umum_bayar = total5050umum_bayar
	obj.Total5050special_bayar = total5050special_bayar
	obj.Total5050kombinasi_bayar = total5050kombinasi_bayar
	obj.Totalmacaukombinasi_bayar = totalmacaukombinasi_bayar
	obj.Totaldasar_bayar = totaldasar_bayar
	obj.Totalshio_bayar = totalshio_bayar
	obj.Totalwin_4d = totalwin_4d
	obj.Totalwin_3d = totalwin_3d
	obj.Totalwin_2d = totalwin_2d
	obj.Totalwin_colokbebas = totalwin_colokbebas
	obj.Totalwin_colokmacau = totalwin_colokmacau
	obj.Totalwin_coloknaga = totalwin_coloknaga
	obj.Totalwin_colokjitu = totalwin_colokjitu
	obj.Totalwin_5050umum = totalwin_5050umum
	obj.Totalwin_5050special = totalwin_5050special
	obj.Totalwin_5050kombinasi = totalwin_5050kombinasi
	obj.Totalwin_macaukombinasi = totalwin_macaukombinasi
	obj.Totalwin_dasar = totalwin_dasar
	obj.Totalwin_shio = totalwin_shio
	obj.Subtotal_bayar = subtotal_bayar
	obj.Subtotal_winner = subtotal_winner
	obj.Total_winlose = total_winlose

	res.Status = fiber.StatusOK
	res.Message = msg
	res.Totalrecord = 0
	res.Record = obj
	res.Time = time.Since(render_page).String()
	return res, nil
}

func _checkpasaran(client_company, pasaran_code string) string {
	var myDays = []string{"minggu", "senin", "selasa", "rabu", "kamis", "jumat", "sabtu"}
	statuspasaran := "ONLINE"

	con := db.CreateCon()
	ctx := context.Background()

	tglnow, _ := goment.New()
	daynow := tglnow.Format("d")
	intVar, _ := strconv.ParseInt(daynow, 0, 8)
	daynowhari := myDays[intVar]

	tbl_trx_keluaran, _, _ := Get_mappingdatabase(client_company)

	sqlpasaran := `SELECT 
		idcomppasaran, nmpasarantogel, 
		jamtutup, jamopen  
		FROM ` + config.DB_VIEW_CLIENT_VIEW_PASARAN + `  
		WHERE idcompany = ? 
		AND idpasarantogel = ? 
	`

	rowpasaran, err := con.QueryContext(ctx, sqlpasaran, client_company, pasaran_code)
	defer rowpasaran.Close()
	helpers.ErrorCheck(err)
	for rowpasaran.Next() {
		var (
			idcomppasaran, nmpasarantogel, jamtutup, jamopen string
			idtrxkeluaran, keluaranperiode, haripasaran      string
		)

		err = rowpasaran.Scan(&idcomppasaran, &nmpasarantogel, &jamtutup, &jamopen)
		helpers.ErrorCheck(err)

		sqlkeluaran := `
			SELECT 
			idtrxkeluaran, keluaranperiode
			FROM ` + tbl_trx_keluaran + `  
			WHERE idcompany = ?
			AND idcomppasaran = ?
			AND keluarantogel = ''
			LIMIT 1
		`
		err := con.QueryRowContext(ctx, sqlkeluaran, client_company, idcomppasaran).Scan(&idtrxkeluaran, &keluaranperiode)
		helpers.ErrorCheck(err)

		sqlpasaranonline := `
			SELECT
				haripasaran
			FROM ` + config.DB_tbl_mst_company_game_pasaran_offline + ` 
			WHERE idcomppasaran = ?
			AND idcompany = ? 
			AND haripasaran = ? 
		`

		errpasaranonline := con.QueryRowContext(ctx, sqlpasaranonline, idcomppasaran, client_company, daynowhari).Scan(&haripasaran)

		if errpasaranonline != sql.ErrNoRows {
			tglskrg := tglnow.Format("YYYY-MM-DD HH:mm:ss")
			jamtutup := tglnow.Format("YYYY-MM-DD") + " " + jamtutup
			jamopen := tglnow.Format("YYYY-MM-DD") + " " + jamopen

			if tglskrg >= jamtutup && tglskrg <= jamopen {
				statuspasaran = "OFFLINE"
			}
		}
	}

	return statuspasaran
}
