	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0
v6:	.word	0

	.text
main:
	li $t1, 1		; v1 -> $t1
	li $t4, 2		; v2 -> $t4
	move $t3, $t4		; v3 -> $t3
	move $t2, $t4		; v4 -> $t2
	sw $t2, v4		; spilled v4, freed $t2
	move $t2, $t3		; v5 -> $t2
	sw $t1, v1		; spilled v1, freed $t1
	add $t1, $t2, $t1	; v6 -> $t1

	; Store variables back into memory
	sw $t1, v6
	sw $t4, v2
	sw $t3, v3
	sw $t2, v5
