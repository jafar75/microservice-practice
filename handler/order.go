package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/google/uuid"
	"github.com/jafar75/microservice-practice/model"
	"github.com/jafar75/microservice-practice/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
};

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID          `json:"customer_id"`
		LineItems  []model.LineItem   `json:"line_items"` 
	};

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest);
		return;
	}

	now := time.Now().UTC();
	
	order := model.Order{
		OrderID: rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems: body.LineItems,
		CreatedAt: &now,
	};

	err := o.Repo.Insert(r.Context(), order);
	if err != nil {
		fmt.Println("failed to insert:", err);
		w.WriteHeader(http.StatusInternalServerError);
		return;
	}

	res, err := json.Marshal(order);
	if err != nil {
		fmt.Println("failed to marshal:", err);
		w.WriteHeader(http.StatusInternalServerError);
		return;
	}

	w.Write(res);
	w.WriteHeader(http.StatusCreated);
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("Cursor");
	if cursorStr == "" {
		cursorStr = "0";
	}

	const decimal = 10;
	const bitSize = 64;

	cursor, err := strconv.ParseInt(cursorStr, decimal, bitSize);
	if err != nil {
		w.WriteHeader(http.StatusBadRequest);
		return;
	}

	const size = 50;
	res, err := o.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: uint(cursor),
		Size: size,
	});
	if err != nil {
		fmt.Println("failed to findAll: ", err);
		w.WriteHeader(http.StatusInternalServerError);
	}

	var response struct {
		Items []model.Order	`json:"items"`
		Next uint64			`json:"next,omitempty"`
	};
	response.Items = res.Orders;
	response.Next = res.Cursor;

	data, err := json.Marshal(response);
	if err != nil {
		fmt.Println("failed to marshal: ", err);
		w.WriteHeader(http.StatusInternalServerError);
		return;
	}
	w.Write(data);

}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oo, err := o.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(oo); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) UpdateById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theOrder, err := o.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if theOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.ShippedAt = &now
	case completedStatus:
		if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.Update(r.Context(), theOrder)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theOrder); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.DeleteByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
