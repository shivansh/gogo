; Test to demonstrate register spilling via next-use heuristic.

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
	li $t4, 2		; v2 -> $t4
	li $t1, -12		; v1 -> $t1
	sw $t1, v1		; spilled v1, freed $t1
	li $t1, 3		; v3 -> $t1
	sw $t4, v2		; spilled v2, freed $t4
	li $t4, 4		; v4 -> $t4
	li $t3, 5		; v5 -> $t3
	move $t3, $t4		; v5 -> $t3
	add $t3, $t4, $t1	; v5 -> $t3
	li $t2, 5		; v6 -> $t2
	sw $t3, v5		; spilled v5, freed $t3
	li $t3, 5		; v7 -> $t3
	j temp

	; Store variables back into memory
	sw $t1, v3
	sw $t4, v4
	sw $t3, v7
	sw $t2, v6
temp:
	li $t1, 1		; v1 -> $t1

	; Store variables back into memory
	sw $t1, v1
unsat:
	li $t1, 4		; v4 -> $t1

	; Store variables back into memory
	sw $t1, v4
