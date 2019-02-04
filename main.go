package main

import (
	"encoding/json"
	"flag"

	"fmt"

	"encoding/hex"

	"crypto/sha256"

	"os"

	"encoding/csv"

	"time"

	"github.com/FactomProject/factom"
)

type Row struct {
	ForumPost
	EntryHash          []byte
	InvalidationReason string
	SortableHash       []byte
	Salt               []byte
}

type ForumPost struct {
	EntryDate int64 `json:"entry_date"`
	PostData  struct {
		EditCount      int    `json:"edit_count"`
		LastEditDate   int64  `json:"last_edit_date"`
		LastEditUserID int    `json:"last_edit_user_id"`
		MessageSha512  string `json:"message_sha512"`
		NodeID         int    `json:"node_id"`
		PostDate       int64  `json:"post_date"`
		ThreadID       int    `json:"thread_id"`
		TitleSha512    string `json:"title_sha512"`
		UserID         int    `json:"user_id"`
	} `json:"post_data"`
	PostLink string `json:"post_link"`
}

func (r *Row) ColumnHeaders() []string {
	return []string{"UserID", "InvalidationReason", "EntryHash", "SortableHash", "PostLink", "PostDate"}
}

func (r *Row) Columns() []string {
	if r.InvalidationReason != "" {
		return []string{
			fmt.Sprintf("%d", r.PostData.UserID),
			r.InvalidationReason,
			"",
			"",
			r.PostLink,
			time.Unix(r.PostData.PostDate, 0).UTC().Format(time.RFC822)}
	}
	return []string{
		fmt.Sprintf("%d", r.PostData.UserID),
		r.InvalidationReason,
		fmt.Sprintf("%x", r.EntryHash),
		fmt.Sprintf("%x", r.SortableHash),
		r.PostLink,
		time.Unix(r.PostData.PostDate, 0).UTC().Format(time.RFC822)}
}

func (r *Row) String() string {
	if r.InvalidationReason != "" {
		return fmt.Sprintf("User: %4d, %10s: %s", r.PostData.UserID, "Invalid", r.InvalidationReason)
	}
	return fmt.Sprintf("User: %4d, %10s: %x", r.PostData.UserID, "SortHash", r.SortableHash)
}

func (r *Row) CalcHash(salt []byte) {
	h := sha256.New()
	h.Write(r.EntryHash)
	h.Write(salt)
	r.SortableHash = h.Sum(nil)
	r.Salt = salt
}

func main() {
	var (
		chainid = flag.String("c", "", "ChainID of post")
		saltHex = flag.String("s", "", "Salt to hash with entryhash. Must be in hex!")
		host    = flag.String("h", "localhost:8088", "Factomd host.")
		csvFile = flag.String("csv", "", "Provide output to csvFile file.")
		//userFile = flag.String("u", "factomize_users.txt", "List of Factomize users")
	)

	flag.Parse()

	factom.SetFactomdServer(*host)

	if *saltHex == "" {
		fmt.Println(usage())
		fmt.Println("Must provide a salt")
		return
	}

	if *chainid == "" {
		fmt.Println(usage())
		fmt.Println("Must provide a chainid")
		return
	}

	salt, err := hex.DecodeString(*saltHex)
	if err != nil {
		panic(err)
	}

	entries, err := factom.GetAllChainEntries(*chainid)
	if err != nil {
		panic(err)
	}

	// Was going to include mapping from user id to username, but holding off
	//users := make(map[int]string)
	//uFile, err := os.OpenFile(*userFile, os.O_RDONLY, 0777)
	//if err != nil {
	//	panic(err)
	//}
	//defer uFile.Close()
	//rd := csv.NewReader(uFile)
	//records, err := rd.ReadAll()
	//if err != nil {
	//	panic(err)
	//}

	accounterUsers := make(map[int]int)

	var posts []*Row

	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		post := new(Row)
		err := json.Unmarshal(e.Content, post)
		if err != nil {
			panic(err)
		}
		post.EntryHash = e.Hash()
		if post.PostData.UserID == 0 {
			continue // Entry that starts chain
		}

		posts = append(posts, post)
	}

	for _, p := range posts {
		if amt, ok := accounterUsers[p.PostData.UserID]; ok {
			p.InvalidationReason = fmt.Sprintf("Post number %d by user. Already in raffle", amt)
			accounterUsers[p.PostData.UserID] = amt + 1
		} else {
			accounterUsers[p.PostData.UserID] = 1
		}
		p.CalcHash(salt)
		// Print to stdout if not to file
		if *csvFile == "" {
			fmt.Println(p.String())
		}
	}

	if *csvFile != "" {
		file, err := os.OpenFile(*csvFile, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Write the header things ontop
		now := time.Now().UTC()
		var tmp Row
		writer := csv.NewWriter(file)
		writer.Write([]string{"TFA 2019 Coin Raffle"})
		writer.Write([]string{
			"Coin raffle for chain:",
			fmt.Sprintf("%s", *chainid),
			"Salt: " + fmt.Sprintf("%s", *saltHex),
			"Time: " + now.Format(time.RFC822),
		})
		writer.Write([]string{})

		// Column headers
		err = writer.Write(tmp.ColumnHeaders())
		if err != nil {
			panic(err)
		}
		for _, r := range posts {
			err = writer.Write(r.Columns())
			if err != nil {
				panic(err)
			}
		}
		writer.Flush()
	}
}

func usage() string {
	return fmt.Sprintf("coin-raffle -s SALT -c CHAIN_ID")
}

/*

{
   "entry_date":1548972114,
   "post_data":{
      "edit_count":0,
      "last_edit_date":0,
      "last_edit_user_id":0,
      "message_sha512":"7841f74a149d8fcefbfd747f5ed57042836d091ec8d2d4d3716f831e8e9b32b32f6cb445f61f6d722379138b07f532eedbb8acdbabb70a516ce6cc4b9be92b48",
      "node_id":52,
      "post_date":1548971254,
      "thread_id":1575,
      "title_sha512":"cad89aeaf32763f5aa72a8a884da5c0363bc69663a02fe355c19bc5a9fb4e4935e78f61307257c3d69bb11786ba118d305aba0b9be7f17c5e34d8b50ca733d72",
      "user_id":9
   },
   "post_link":"https://factomize.com/forums/index.php?threads/1575#post-10623"
}
*/
