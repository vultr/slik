package labeler

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type Labels struct {
	NodeName        string
	CPUs            int
	Boards          int
	SocketsPerBoard int
	CoresPerSocket  int
	ThreadsPerCore  int
	RealMemory      int
}

// NewLabeler returns a labeler based on slurmd -C output
func NewLabeler(out string) *Labels {
	log := zap.L().Sugar()

	log.Debug(out)

	var labels Labels

	fields := strings.Fields(out)

	for i := range fields {
		entry := strings.Split(fields[i], "=")

		log.Debugf("%+v", entry)

		if len(entry) != 2 {
			log.Info("entry not length of 2, skipping...")

			continue
		}

		key := entry[0]
		val := entry[1]

		switch key {
		case "NodeName":
			labels.NodeName = os.Getenv("HOSTNAME")
		case "CPUs":
			cpus, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.CPUs = cpus
		case "Boards":
			boards, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.Boards = boards
		case "SocketsPerBoard":
			sockets, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.SocketsPerBoard = sockets
		case "CoresPerSocket":
			cores, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.CoresPerSocket = cores
		case "ThreadsPerCore":
			threads, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.ThreadsPerCore = threads
		case "RealMemory":
			mem, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err)

				continue
			}

			labels.RealMemory = mem
		}
	}

	return &labels
}
