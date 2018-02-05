	.data
v1:	.word	0
v2:	.word	0

	.text
main:
	li $t1, 1		; v1 -> $t1
	li $t2, 2		; v2 -> $t2
	li $t1, 4

	; Store variables back into memory
	sw $t2, ($t0)
	sw $t1, ($t0)
