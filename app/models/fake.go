package models

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func fakeCreateUser(handle string, privilege int8, bio string) {
	u := User{
		Handle:    handle,
		Email:     handle + "@example.com",
		Password:  "qwq",
		Privilege: privilege,
		Nickname:  "~ " + handle + " ~",
		Bio:       bio,
	}
	if err := u.Create(); err != nil {
		panic(err)
	}
	log.Println("User " + handle + " created")
}

func FakeDatabase() {
	// Clear database
	for _, schema := range schemata {
		_, err := db.Exec("DROP TABLE IF EXISTS " + schema.table + " CASCADE")
		if err != nil {
			panic(err)
		}
	}
	InitializeSchemata(db)

	// Clear Redis
	if rcli != nil {
		_, err := rcli.FlushDB().Result()
		if err != nil {
			panic(err)
		}
		InitializeRedis(rcli)
	}

	// Users
	// - Superuser
	fakeCreateUser("su", UserPrivilegeSuperuser, "I have been notified")
	// - Organizers
	for i := 1; i <= 5; i++ {
		fakeCreateUser("o"+strconv.Itoa(i), UserPrivilegeOrganizer, "Enjoy the contests")
	}
	// - Participants
	for i := 1; i <= 20; i++ {
		fakeCreateUser("p"+strconv.Itoa(i), UserPrivilegeNormal, "I'm a teapot")
	}

	// Contests
	t := time.Now().Unix()
	numbers := []string{"zero", "one", "two", "three", "four", "five"}
	for i := 1; i <= 5; i++ {
		s := "This is the description for contest number " + numbers[i] + "!\n"
		s += "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris. Integer in mauris eu nibh euismod gravida. Duis ac tellus et risus vulputate vehicula. Donec lobortis risus a elit. Etiam tempor. Ut ullamcorper, ligula eu tempor congue, eros est euismod turpis, id tincidunt sapien risus a quam. Maecenas fermentum consequat mi. Donec fermentum. Pellentesque malesuada nulla a mi. Duis sapien sem, aliquet nec, commodo eget, consequat quis, neque. Aliquam faucibus, elit ut dictum aliquet, felis nisl adipiscing sapien, sed malesuada diam lacus eget erat. Cras mollis scelerisque nunc. Nullam arcu. Aliquam consequat. Curabitur augue lorem, dapibus quis, laoreet et, pretium ac, nisi. Aenean magna nisl, mollis quis, molestie eu, feugiat in, orci. In hac habitasse platea dictumst."
		script := `
local count = 9
local su_id = get_id('su')

function on_submission(all, from)
    print('Submission', from)
    for i = 1, #all do
        print(all[i], all[i] == from)
    end
end

function on_timer(all)
    count = count + 1
    if count < 10 then return end
    count = 0
    print('Superuser has ID ' .. tostring(su_id))
    print('Creating matches for contest #` + strconv.Itoa(i) + `')
    print('Number of participants with delegates ' .. tostring(#all))
    for i = 1, #all do
        print(string.format('Contestant %s (%d)', get_handle(all[i]), all[i]))
        if i > 1 then create_match(all[i], all[i - 1]) end
    end
end

function on_manual(all, arg)
    print('Manual', arg)
end

function update_stats(report, par)
    print('Update with ' .. tostring(#par) .. ' parties')
    print(report)
    for i = 1, #par do
        print(i, par[i].rating, par[i].performance)
        par[i].rating = par[i].rating + 1
        par[i].performance = 'hahaha'
    end
end
`
		c := Contest{
			Title:     "Grand Contest " + strconv.Itoa(i),
			Banner:    "banner.png",
			Owner:     int32(1 + i),
			StartTime: t + 3600*24*int64(-3+i),
			EndTime:   t + 3600*24*int64(-1+i),
			Desc:      "Really big contest, number " + numbers[i],
			Details:   s,
			IsVisible: i != 1,
			IsRegOpen: i != 5,
			Script:    script,
			Playback:  "<html><body><p>Report:</p><pre><% report %></pre></body></html>",
		}
		if err := c.Create(); err != nil {
			panic(err)
		}

		// Judge
		judgeCode, err := ioutil.ReadFile("../ipc/run.c")
		if err != nil {
			panic(err)
		}
		j := Submission{
			User:     int32(6 + (1 + i/2)),
			Contest:  int32(i),
			Language: "c",
			Contents: string(judgeCode),
		}
		if err := j.Create(); err != nil {
			panic(err)
		}
		j.SendToQueue()

		c.Judge = j.Id
		if err := c.Update(); err != nil {
			panic(err)
		}

		// Participants
		for j := 1 + i/2; j <= 20; j += i {
			log.Printf("User %d joins contest %d\n", j, i)
			p := ContestParticipation{
				User:    int32(6 + j),
				Contest: int32(i),
				Type:    ParticipationTypeContestant,
			}
			if err := p.Create(); err != nil {
				panic(err)
			}

			// Submissions
			for k := 1; k <= 2+(i+j)%3; k++ {
				s := Submission{
					User:     int32(6 + j),
					Contest:  int32(i),
					Language: "c",
					Contents: `
#include "ipc.h"

#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>
#include <time.h>

int main()
{
    char *s = ipc_recv_str();
    int side;
    sscanf(s, "%d", &side);
    free(s);
    fprintf(stderr, "Hello, submission ` + strconv.Itoa(rand.Intn(900000)+100000) + ` from side #%d\n", side);

    srand(((unsigned)time(NULL) << 1) | side);
    bool board[3][3] = {{ false }};

    while (1) {
        // Board state change
        int row, col;
        s = ipc_recv_str();
        sscanf(s, "%d%d", &row, &col);
        free(s);
        if (row != -1) board[row][col] = true;

        // Pick a random cell
        int u, v;
        do {
            u = rand() % 3;
            v = rand() % 3;
        } while (board[u][v]);
        board[u][v] = true;
        fprintf(stderr, "Moving at (%d, %d)\n", u, v);

        // Send
        char t[8];
        sprintf(t, "%d %d", u, v);
        ipc_send_str(t);
    }

    return 0;
}
`,
				}
				if err := s.Create(); err != nil {
					panic(err)
				}
				// TODO: Move delegate & match creation to a separate endpoint
				p.Delegate = s.Id
				s.SendToQueue()
			}

			if err := p.Update(); err != nil {
				panic(err)
			}
		}
	}
}

func FakeMatches() {
	_, err := db.Exec("UPDATE match SET status = $1", MatchStatusDone)
	if err != nil {
		panic(err)
	}
}
