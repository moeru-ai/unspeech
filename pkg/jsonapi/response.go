package jsonapi

type Response struct {
	// Data is the primary data for a response.
	Data []any `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
	// Errors is an array of error objects.
	Errors []*ErrorObject `protobuf:"bytes,2,rep,name=errors,proto3" json:"errors,omitempty"`
}
