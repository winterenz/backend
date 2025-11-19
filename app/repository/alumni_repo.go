package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlumniRepository interface {
	List(ctx context.Context) ([]model.Alumni, error)
	GetByID(ctx context.Context, id string) (*model.Alumni, error)
	GetByUserID(ctx context.Context, userID string) (*model.Alumni, error)
	Create(ctx context.Context, in model.CreateAlumniReq, userID string) (string, error)
	Update(ctx context.Context, id string, in model.UpdateAlumniReq) error
	Delete(ctx context.Context, id string) error
	ListByJurusan(ctx context.Context, jurusan string) ([]model.Alumni, error)
	ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Alumni, error)
	Count(ctx context.Context, search string) (int64, error)
}

type alumniRepo struct {
	c *mongo.Collection
}

func NewAlumniRepository(db *mongo.Database) AlumniRepository {
	return &alumniRepo{c: db.Collection("alumni")}
}

func toOID(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}

var alumniSortCols = map[string]string{
	"id":           "_id",
	"nim":          "nim",
	"nama":         "nama",
	"jurusan":      "jurusan",
	"angkatan":     "angkatan",
	"tahun_lulus":  "tahun_lulus",
	"email":        "email",
	"created_at":   "created_at",
	"updated_at":   "updated_at",
}

func normalizeSort(col, order string) bson.D {
	field, ok := alumniSortCols[strings.ToLower(col)]
	if !ok {
		field = "_id"
	}
	dir := int32(1)
	if strings.EqualFold(order, "desc") {
		dir = -1
	}
	return bson.D{{Key: field, Value: dir}}
}

func (r *alumniRepo) List(ctx context.Context) ([]model.Alumni, error) {
	cur, err := r.c.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Alumni
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *alumniRepo) GetByID(ctx context.Context, id string) (*model.Alumni, error) {
	oid, err := toOID(id)
	if err != nil {
		return nil, err
	}
	var a model.Alumni
	err = r.c.FindOne(ctx, bson.M{"_id": oid}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &a, err
}

func (r *alumniRepo) GetByUserID(ctx context.Context, userID string) (*model.Alumni, error) {
	uid, err := toOID(userID)
	if err != nil {
		return nil, err
	}
	var a model.Alumni
	err = r.c.FindOne(ctx, bson.M{"user_id": uid}).Decode(&a)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &a, err
}

func (r *alumniRepo) Create(ctx context.Context, in model.CreateAlumniReq, userID string) (string, error) {
	uid, err := toOID(userID)
	if err != nil {
		return "", err
	}
	now := time.Now()
	doc := model.Alumni{
		ID:         primitive.NewObjectID(),
		UserID:     uid,
		NIM:        in.NIM,
		Nama:       in.Nama,
		Jurusan:    in.Jurusan,
		Angkatan:   in.Angkatan,
		TahunLulus: in.TahunLulus,
		Email:      in.Email,
		NoTelepon:  in.NoTelepon,
		Alamat:     in.Alamat,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := r.c.InsertOne(ctx, doc); err != nil {
		return "", err
	}
	return doc.ID.Hex(), nil
}

func (r *alumniRepo) Update(ctx context.Context, id string, in model.UpdateAlumniReq) error {
	oid, err := toOID(id)
	if err != nil {
		return err
	}

	set := bson.M{
		"updated_at": time.Now(),
	}

	if in.NIM != nil {
		set["nim"] = *in.NIM
	}
	if in.Nama != nil {
		set["nama"] = *in.Nama
	}
	if in.Jurusan != nil {
		set["jurusan"] = *in.Jurusan
	}
	if in.Angkatan != nil {
		set["angkatan"] = *in.Angkatan
	}
	if in.TahunLulus != nil {
		set["tahun_lulus"] = *in.TahunLulus
	}
	if in.Email != nil {
		set["email"] = *in.Email
	}
	if in.NoTelepon != nil {
		set["no_telepon"] = *in.NoTelepon
	}
	if in.Alamat != nil {
		set["alamat"] = *in.Alamat
	}
	if len(set) == 1 { // hanya updated_at
		return errors.New("tidak ada field yang di-update")
	}

	res, err := r.c.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *alumniRepo) Delete(ctx context.Context, id string) error {
	oid, err := toOID(id)
	if err != nil {
		return err
	}
	res, err := r.c.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *alumniRepo) ListByJurusan(ctx context.Context, jurusan string) ([]model.Alumni, error) {
	filter := bson.M{"jurusan": jurusan}
	cur, err := r.c.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Alumni
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *alumniRepo) ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Alumni, error) {
	filter := bson.M{}
	if strings.TrimSpace(search) != "" {
		filter["$or"] = []bson.M{
			{"nama":    bson.M{"$regex": search, "$options": "i"}},
			{"email":   bson.M{"$regex": search, "$options": "i"}},
			{"jurusan": bson.M{"$regex": search, "$options": "i"}},
			{"nim":     bson.M{"$regex": search, "$options": "i"}},
			{"alamat":  bson.M{"$regex": search, "$options": "i"}},
		}
	}

	opts := options.Find().
		SetSort(normalizeSort(sortBy, order)).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cur, err := r.c.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Alumni
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *alumniRepo) Count(ctx context.Context, search string) (int64, error) {
	filter := bson.M{}
	if strings.TrimSpace(search) != "" {
		filter["$or"] = []bson.M{
			{"nama":    bson.M{"$regex": search, "$options": "i"}},
			{"email":   bson.M{"$regex": search, "$options": "i"}},
			{"jurusan": bson.M{"$regex": search, "$options": "i"}},
			{"nim":     bson.M{"$regex": search, "$options": "i"}},
			{"alamat":  bson.M{"$regex": search, "$options": "i"}},
		}
	}
	return r.c.CountDocuments(ctx, filter)
}
