p = input().strip()
s = input().strip()

words = s.split()

list = {}
for char, word in zip(p, words):
    if char in list:
        if list[char] != word:
            print("No")
            break
    else:
        list[char] = word
else:
    print("Yes")