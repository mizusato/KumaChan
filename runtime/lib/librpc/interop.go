package librpc

import (
	"io"
	"bytes"
	. "kumachan/lang"
	"kumachan/rpc"
	"kumachan/rpc/kmd"
)


type RpcApiImpl struct {
	kmdApi  KmdApi
	index   rpc.ServiceIndex
}
type RpcInfoContext interface {
	KmdTransformContext
	GetRpcInfo()  RpcInfo
}
func CreateRpcApi(ctx RpcInfoContext) RpcApi {
	return &RpcApiImpl {
		kmdApi: CreateKmdApi(ctx),
		index:  ctx.GetRpcInfo().ServiceIndex,
	}
}
func (impl *RpcApiImpl) GetKmdApi() KmdApi {
	return impl.kmdApi
}
func (impl *RpcApiImpl) GetServiceInterface(id rpc.ServiceIdentifier) (rpc.ServiceInterface, bool) {
	var i, exists = impl.index[id]
	return i, exists
}

type KmdApiImpl struct {
	config      KmdInfo
	transformer kmd.Transformer
}
type KmdTransformContext interface {
	KmdGetInfo() KmdInfo
	KmdGetAdapter(index uint) Value
	KmdCallAdapter(f Value, x Value) Value
	KmdCallValidator(f Value, x Value) bool
}
func CreateKmdApi(ctx KmdTransformContext) KmdApi {
	return &KmdApiImpl {
		config:      ctx.KmdGetInfo(),
		transformer: kmdCreateTransformer(ctx),
	}
}
func (impl *KmdApiImpl) GetTypeFromId(id kmd.TypeId) *kmd.Type {
	return impl.config.GetTypeFromId(id)
}
func (impl *KmdApiImpl) SerializeToStream(v Value, t *kmd.Type, stream io.Writer) error {
	var serializer = impl.transformer.Serializer
	var tv = KmdTypedValue {
		Type:  t,
		Value: v,
	}
	return kmd.Serialize(tv, serializer, stream)
}
func (impl *KmdApiImpl) DeserializeFromStream(t *kmd.Type, stream io.Reader) (Value, error) {
	var ts = impl.transformer
	var deserializer = ts.Deserializer
	obj, real_t, err := kmd.Deserialize(stream, deserializer)
	if err != nil { return nil, err }
	obj, err = ts.AssignObject(obj, real_t, t)
	if err != nil { return nil, err }
	return obj, nil
}
func (impl *KmdApiImpl) Serialize(v Value, t *kmd.Type) ([] byte, error) {
	var buf bytes.Buffer
	var err = impl.SerializeToStream(v, t, &buf)
	if err != nil { return nil, err }
	return buf.Bytes(), nil
}
func (impl *KmdApiImpl) Deserialize(binary ([] byte), t *kmd.Type) (Value, error) {
	var reader = bytes.NewReader(binary)
	return impl.DeserializeFromStream(t, reader)
}

