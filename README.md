# Coin Raffle
To try this yourself, choose a thread on factomize.com.
1. Locate it's chain id on the Factom blockchain. (All posts are factomized to factom blockchain)
2. Choose a salt (TFA will choose the salt of our post that closes the raffle)
3. Run the command line app against a factomd node (localhost:8088 is default)
4. Choose an output csv file

```
coin-raffle -c CHAIN_ID -s SALT -csv out.csv -h FACTOM_NODE
```

Open the CSV file in excel or google sheets. Freeze the top 4 rows (these are just headers). Sort the "SortableHash"
column from A-Z. This is the order of the winners!