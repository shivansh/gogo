	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0
v6:	.word	0
v7:	.word	0

	.text
main:
	li $t1, -1		; v1 -> $t1
	li $t2, 2		; v2 -> $t2
	li $t1, -12		; v1 -> $t1
	li $t3, 3		; v3 -> $t3
	li $t4, 4		; v4 -> $t4
	sw $t1, v1		; spilled v1, freed $t1
	li $t1, 5		; v5 -> $t1
	move $t1, $t4		; v5 -> $t1
	add $t1, $t4, $t3	; v5 -> $t1
	sw $t1, v5		; spilled v5, freed $t1
	li $t1, 5		; v6 -> $t1
	sw $t1, v6		; spilled v6, freed $t1
	li $t1, 5		; v7 -> $t1
	j temp

	; Store variables back into memory
	sw $t4, v4
	sw $t1, v7
	sw $t2, v2
	sw $t3, v3
temp:

	; Store variables back into memory
unsat:
	li $t1, 4		; v4 -> $t1

	; Store variables back into memory
	sw $t1, v4
