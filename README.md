# multi-key-pprlgo-for-frozen-lake

## How to run

1. git clone git@github.com:Kizitora3000/multi-key-pprlgo-for-frozen-lake.git
2. cd multi-key-pprlgo-for-frozen-lake
3. go get MKpprlgoFrozenLake/mkckks
4. go mod download github.com/ldsec/lattigo/v2
5. go run main.go -s 4x4
    + -s: map size (4x4 or 5x5 or 6x6)

## Setup paramerters

edit main.go

+ EPISODES: 学習を完了するまでのエピソード数 (論文: 200)
+ MAX_USERS: 学習に参加するユーザ数 (論文: 1 to 3)
+ MAX_TRIALS: 試行回数 (論文: 100)
+ ckks parameters: NewParametersFromLiteral(utils.FAST_BUT_NOT_128) (論文: utils.PN15QP880 (非常に時間かかる))