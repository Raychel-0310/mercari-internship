class ListNode:
    def __init__(self, val=0, next=None):
        self.val = val
        self.next = next

def get_node(headA, headB):
    A, B = headA, headB
    while A != B:
        A = A.next if A else headB
        B = B.next if B else headA
    return A