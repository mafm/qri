package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/qri-io/qri/repo"

	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
)

// MtDatasets is a dataset list message
const MtDatasets = MsgType("list_datasets")

// listMax is the highest number of entries a list request should return
const listMax = 30

// DatasetsListParams encapsulates options for requesting datasets
type DatasetsListParams struct {
	Limit  int
	Offset int
}

// RequestDatasetsList gets a list of a peer's datasets
func (n *QriNode) RequestDatasetsList(pid peer.ID, p DatasetsListParams) ([]repo.DatasetRef, error) {
	log.Debugf("%s RequestDatasetList: %s", n.ID, pid)

	if pid == n.ID {
		// requesting self isn't a network operation
		return n.Repo.References(p.Limit, p.Offset)
	}

	req, err := NewJSONBodyMessage(n.ID, MtDatasets, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	req = req.WithHeaders("phase", "request")

	replies := make(chan Message)
	err = n.SendMessage(req, replies, pid)
	if err != nil {
		log.Debug(err.Error())
		return nil, fmt.Errorf("send dataset info message error: %s", err.Error())
	}

	res := <-replies
	ref := []repo.DatasetRef{}
	err = json.Unmarshal(res.Body, &ref)
	return ref, err
}

func (n *QriNode) handleDatasetsList(ws *WrappedStream, msg Message) (hangup bool) {
	hangup = true
	switch msg.Header("phase") {
	case "request":
		dlp := DatasetsListParams{}
		if err := json.Unmarshal(msg.Body, &dlp); err != nil {
			log.Debugf("%s %s", n.ID, err.Error())
			return
		}

		if dlp.Limit == 0 || dlp.Limit > listMax {
			dlp.Limit = listMax
		}

		refs, err := n.Repo.References(dlp.Limit, dlp.Offset)
		if err != nil {
			log.Debug(err.Error())
			return
		}

		// replies := make([]*repo.DatasetRef, p.Limit)
		// i := 0
		// for i, ref := range refs {
		// 	if i >= p.Limit {
		// 		break
		// 	}
		// 	ds, err := dsfs.LoadDataset(n.Repo.Store(), datastore.NewKey(ref.Path))
		// 	if err != nil {
		// 		log.Info("error loading dataset at path:", ref.Path)
		// 		return nil
		// 	}
		// 	refs[i].Dataset = ds
		// 	// i++
		// }

		reply, err := msg.UpdateJSON(refs)
		reply = reply.WithHeaders("phase", "response")
		if err := ws.sendMessage(reply); err != nil {
			log.Debug(err.Error())
			return
		}
	}

	return
}

// func (n *QriNode) handleDatasetsResponse(pi pstore.PeerInfo, r *Message) error {
// 	data, err := json.Marshal(r.Body)
// 	if err != nil {
// 		return err
// 	}
// 	ds := []*repo.DatasetRef{}
// 	if err := json.Unmarshal(data, &ds); err != nil {
// 		return err
// 	}
// 	return n.Repo.Cache().PutDatasets(ds)
// }

// MtDatasetsCreated announces the creation of one or more datasets
// const MtDatasetsCreated = MsgType("datasets_created")

// func (n *QriNode) AnnounceDatasetsCreated(ds ...repo.DatasetRef) error {

// }

// func (n QriNode) handleDatasestsCreated(ws *WrappedStream, msg Message) error {

// }

// // MtDatasetInfo gets info on a dataset
// const MtDatasetsDeleted = MsgType("datasets_deleted")

// func (n *QriNode) AnnounceDatasetsCreated(ds ...repo.DatasetRef) error {

// }

// func (n QriNode) handleDatasestsCreated(ws *WrappedStream, msg Message) error {

// }
