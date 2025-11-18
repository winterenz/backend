package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"prak3/clean-architecture-fiber-mongo/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PekerjaanRepository interface {
	// read
	List(ctx context.Context) ([]model.Pekerjaan, error)
	GetByID(ctx context.Context, id string) (*model.Pekerjaan, error)
	GetByIDAny(ctx context.Context, id string) (*model.Pekerjaan, error) // termasuk yang soft-deleted
	ListByAlumniID(ctx context.Context, alumniID string) ([]model.Pekerjaan, error)

	// write
	Create(ctx context.Context, in model.CreatePekerjaanReq) (string, error)
	Update(ctx context.Context, id string, in model.UpdatePekerjaanReq) error

	// pagination
	ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error)
	Count(ctx context.Context, search string) (int64, error)

	// soft delete
	SoftDeleteAdmin(ctx context.Context, id string, username string) error
	SoftDeleteOwned(ctx context.Context, id string, alumniID string, username string) error

	// trash list
	ListTrashed(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error)
	CountTrashed(ctx context.Context, search string) (int64, error)

	// restore
	RestoreAdmin(ctx context.Context, id string, username string) error
	RestoreOwned(ctx context.Context, id string, alumniID string, username string) error

	// hard delete
	HardDeleteAdmin(ctx context.Context, id string) error
	HardDeleteOwned(ctx context.Context, id string, alumniID string) error

	// validasi kepemilikan
	GetAlumniByUserID(ctx context.Context, userID string) (*model.Alumni, error)
}

type pekerjaanRepo struct {
	cPekerjaan *mongo.Collection
	cAlumni    *mongo.Collection
}

func NewPekerjaanRepository(db *mongo.Database) PekerjaanRepository {
	return &pekerjaanRepo{
		cPekerjaan: db.Collection("pekerjaan_alumni"),
		cAlumni:    db.Collection("alumni"),
	}
}

// ---------- helpers ----------
func oid(hex string) (primitive.ObjectID, error) { return primitive.ObjectIDFromHex(hex) }

var pekerjaanSortCols = map[string]string{
	"id": "_id", "alumni_id": "alumni_id",
	"nama_perusahaan": "nama_perusahaan",
	"posisi_jabatan":  "posisi_jabatan",
	"bidang_industri": "bidang_industri",
	"lokasi_kerja":    "lokasi_kerja",
	"gaji_range":      "gaji_range",
	"tanggal_mulai_kerja":  "tanggal_mulai_kerja",
	"tanggal_selesai_kerja": "tanggal_selesai_kerja",
	"status_pekerjaan": "status_pekerjaan",
	"deskripsi_pekerjaan": "deskripsi_pekerjaan",
	"created_at": "created_at", "updated_at": "updated_at",
}

func normalizePekSort(col, order string) bson.D {
	f, ok := pekerjaanSortCols[strings.ToLower(col)]
	if !ok { f = "_id" }
	dir := int32(1)
	if strings.EqualFold(order, "desc") { dir = -1 }
	return bson.D{{Key: f, Value: dir}}
}

func searchFilter(base bson.M, search string) bson.M {
	if strings.TrimSpace(search) == "" { return base }
	base["$or"] = []bson.M{
		{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
		{"posisi_jabatan":  bson.M{"$regex": search, "$options": "i"}},
		{"bidang_industri": bson.M{"$regex": search, "$options": "i"}},
		{"lokasi_kerja":    bson.M{"$regex": search, "$options": "i"}},
		{"status_pekerjaan":bson.M{"$regex": search, "$options": "i"}},
		{"deskripsi_pekerjaan": bson.M{"$regex": search, "$options": "i"}},
	}
	return base
}

// ---------- reads ----------
func (r *pekerjaanRepo) List(ctx context.Context) ([]model.Pekerjaan, error) {
	cur, err := r.cPekerjaan.Find(ctx, bson.M{"is_deleted": false}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil { return nil, err }
	defer cur.Close(ctx)
	var out []model.Pekerjaan
	if err := cur.All(ctx, &out); err != nil { return nil, err }
	return out, nil
}

func (r *pekerjaanRepo) GetByID(ctx context.Context, id string) (*model.Pekerjaan, error) {
	_id, err := oid(id); if err != nil { return nil, err }
	var p model.Pekerjaan
	err = r.cPekerjaan.FindOne(ctx, bson.M{"_id": _id, "is_deleted": false}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) { return nil, nil }
	return &p, err
}

func (r *pekerjaanRepo) GetByIDAny(ctx context.Context, id string) (*model.Pekerjaan, error) {
	_id, err := oid(id); if err != nil { return nil, err }
	var p model.Pekerjaan
	err = r.cPekerjaan.FindOne(ctx, bson.M{"_id": _id}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) { return nil, nil }
	return &p, err
}

func (r *pekerjaanRepo) ListByAlumniID(ctx context.Context, alumniID string) ([]model.Pekerjaan, error) {
	aid, err := oid(alumniID); if err != nil { return nil, err }
	cur, err := r.cPekerjaan.Find(ctx, bson.M{"alumni_id": aid, "is_deleted": false}, options.Find().SetSort(bson.D{{Key: "tanggal_mulai_kerja", Value: -1}}))
	if err != nil { return nil, err }
	defer cur.Close(ctx)
	var out []model.Pekerjaan
	if err := cur.All(ctx, &out); err != nil { return nil, err }
	return out, nil
}

// ---------- writes ----------
func (r *pekerjaanRepo) Create(ctx context.Context, in model.CreatePekerjaanReq) (string, error) {
	aid, err := oid(in.AlumniID); if err != nil { return "", err }
	now := time.Now()
	doc := model.Pekerjaan{
		ID:                 primitive.NewObjectID(),
		AlumniID:           aid,
		NamaPerusahaan:     in.NamaPerusahaan,
		PosisiJabatan:      in.PosisiJabatan,
		BidangIndustri:     in.BidangIndustri,
		LokasiKerja:        in.LokasiKerja,
		GajiRange:          in.GajiRange,
		TanggalMulaiKerja:  in.TanggalMulaiKerja,
		TanggalSelesaiKerja: in.TanggalSelesaiKerja,
		StatusPekerjaan:    in.StatusPekerjaan,
		Deskripsi:          in.Deskripsi,
		CreatedAt:          now,
		UpdatedAt:          now,
		IsDeleted:          false,
	}
	if _, err := r.cPekerjaan.InsertOne(ctx, doc); err != nil { return "", err }
	return doc.ID.Hex(), nil
}

func (r *pekerjaanRepo) Update(ctx context.Context, id string, in model.UpdatePekerjaanReq) error {
	_id, err := oid(id); if err != nil { return err }
	set := bson.M{
		"nama_perusahaan": in.NamaPerusahaan,
		"posisi_jabatan":  in.PosisiJabatan,
		"bidang_industri": in.BidangIndustri,
		"lokasi_kerja":    in.LokasiKerja,
		"gaji_range":      in.GajiRange,
		"tanggal_mulai_kerja":  in.TanggalMulaiKerja,
		"tanggal_selesai_kerja": in.TanggalSelesaiKerja,
		"status_pekerjaan": in.StatusPekerjaan,
		"deskripsi_pekerjaan": in.Deskripsi,
		"updated_at":      time.Now(),
	}
	// alumni_id boleh diubah (opsional) — kalau body kirim AlumniID
	if strings.TrimSpace(in.AlumniID) != "" {
		if aid, err := oid(in.AlumniID); err == nil {
			set["alumni_id"] = aid
		}
	}
	res, err := r.cPekerjaan.UpdateOne(ctx, bson.M{"_id": _id, "is_deleted": false}, bson.M{"$set": set})
	if err != nil { return err }
	if res.MatchedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

// ---------- pagination (non-trash) ----------
func (r *pekerjaanRepo) ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error) {
	filter := searchFilter(bson.M{"is_deleted": false}, search)
	opts := options.Find().
		SetSort(normalizeSort(sortBy, order)).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	cur, err := r.cPekerjaan.Find(ctx, filter, opts)
	if err != nil { return nil, err }
	defer cur.Close(ctx)
	var out []model.Pekerjaan
	if err := cur.All(ctx, &out); err != nil { return nil, err }
	return out, nil
}

func (r *pekerjaanRepo) Count(ctx context.Context, search string) (int64, error) {
	filter := searchFilter(bson.M{"is_deleted": false}, search)
	return r.cPekerjaan.CountDocuments(ctx, filter)
}

// ---------- soft delete ----------
func (r *pekerjaanRepo) SoftDeleteAdmin(ctx context.Context, id string, username string) error {
	_id, err := oid(id); if err != nil { return err }
	set := bson.M{"is_deleted": true, "deleted_by": username, "deleted_at": time.Now()}
	res, err := r.cPekerjaan.UpdateOne(ctx, bson.M{"_id": _id, "is_deleted": false}, bson.M{"$set": set})
	if err != nil { return err }
	if res.MatchedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

func (r *pekerjaanRepo) SoftDeleteOwned(ctx context.Context, id string, alumniID string, username string) error {
	_id, err := oid(id); if err != nil { return err }
	aid, err := oid(alumniID); if err != nil { return err }
	set := bson.M{"is_deleted": true, "deleted_by": username, "deleted_at": time.Now()}
	filter := bson.M{"_id": _id, "alumni_id": aid, "is_deleted": false}
	res, err := r.cPekerjaan.UpdateOne(ctx, filter, bson.M{"$set": set})
	if err != nil { return err }
	if res.MatchedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

// ---------- trash list ----------
func (r *pekerjaanRepo) ListTrashed(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error) {
	filter := searchFilter(bson.M{"is_deleted": true}, search)
	opts := options.Find().
		SetSort(normalizeSort(sortBy, order)).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	cur, err := r.cPekerjaan.Find(ctx, filter, opts)
	if err != nil { return nil, err }
	defer cur.Close(ctx)
	var out []model.Pekerjaan
	if err := cur.All(ctx, &out); err != nil { return nil, err }
	return out, nil
}

func (r *pekerjaanRepo) CountTrashed(ctx context.Context, search string) (int64, error) {
	filter := searchFilter(bson.M{"is_deleted": true}, search)
	return r.cPekerjaan.CountDocuments(ctx, filter)
}

// ---------- restore ----------
func (r *pekerjaanRepo) RestoreAdmin(ctx context.Context, id string, _ string) error {
	_id, err := oid(id); if err != nil { return err }
	res, err := r.cPekerjaan.UpdateOne(ctx, bson.M{"_id": _id, "is_deleted": true},
		bson.M{"$set": bson.M{"is_deleted": false, "updated_at": time.Now()}, "$unset": bson.M{"deleted_by": "", "deleted_at": ""}})
	if err != nil { return err }
	if res.MatchedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

func (r *pekerjaanRepo) RestoreOwned(ctx context.Context, id string, alumniID string, _ string) error {
	_id, err := oid(id); if err != nil { return err }
	aid, err := oid(alumniID); if err != nil { return err }
	filter := bson.M{"_id": _id, "alumni_id": aid, "is_deleted": true}
	res, err := r.cPekerjaan.UpdateOne(ctx, filter,
		bson.M{"$set": bson.M{"is_deleted": false, "updated_at": time.Now()}, "$unset": bson.M{"deleted_by": "", "deleted_at": ""}})
	if err != nil { return err }
	if res.MatchedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

// ---------- hard delete ----------
func (r *pekerjaanRepo) HardDeleteAdmin(ctx context.Context, id string) error {
	_id, err := oid(id); if err != nil { return err }
	res, err := r.cPekerjaan.DeleteOne(ctx, bson.M{"_id": _id, "is_deleted": true})
	if err != nil { return err }
	if res.DeletedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

func (r *pekerjaanRepo) HardDeleteOwned(ctx context.Context, id string, alumniID string) error {
	_id, err := oid(id); if err != nil { return err }
	aid, err := oid(alumniID); if err != nil { return err }
	res, err := r.cPekerjaan.DeleteOne(ctx, bson.M{"_id": _id, "alumni_id": aid, "is_deleted": true})
	if err != nil { return err }
	if res.DeletedCount == 0 { return mongo.ErrNoDocuments }
	return nil
}

// ---------- helper: ambil alumni milik user ----------
func (r *pekerjaanRepo) GetAlumniByUserID(ctx context.Context, userID string) (*model.Alumni, error) {
	uid, err := oid(userID); if err != nil { return nil, err }
	var a model.Alumni
	err = r.cAlumni.FindOne(ctx, bson.M{"user_id": uid}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) { return nil, nil }
	return &a, err
}
