	.data
v1:	.word	0
v2:	.word	0

	.text
main:
	li $t1, 1		; v1 -> $t1
	li $t4, 2		; v2 -> $t4
	li $t1, 4		; v1 -> $t1

	; Store variables back into memory
	sw $t1, v1
	sw $t4, v2
