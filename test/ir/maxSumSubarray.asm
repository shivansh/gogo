	.data
a:	.space	40
sum:	.space	40
i:	.word	0
v:	.word	0
v1:	.word	0
v2:	.word	0
maxsum:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, 0		# i -> $t1
	# Store variables back into memory
	sw $t1, i

loop1:

	lw $t1, i
	bge $t1, 10, exit1		# exit1 -> $t0
	li $v0, 5
	syscall
	move $t4, $v0
	la $t3, a
	sll $s2, $t1, 2	# iterator *= 4
	sw $t4, a($s2)		# variable -> array
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i
	sw $t4, v

	j loop1

exit1:
	la $t1, a
	lw $t4, 0($t1)		# variable <- array
	la $t3, sum
	sw $t4, 0($t3)		# variable -> array
	li $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i
	sw $t4, v

loop2:

	lw $t1, i
	bge $t1, 10, exit2		# exit2 -> $t0
	sub $t4, $t1, 1		# v1 -> $t4
	# Store variables back into memory
	sw $t1, i
	sw $t4, v1

	lw $t1, v
	bge $t1, 0, branch1		# branch1 -> $t0
	la $t4, a
	sw $t1, v		# spilled v, freed $t1
	lw $t1, i
	sll $s2, $t1, 2	# iterator *= 4
	lw $t3, a($s2)		# variable <- array
	la $t2, sum
	sll $s2, $t1, 2	# iterator *= 4
	sw $t3, sum($s2)		# variable -> array
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i
	sw $t3, v

	j loop2

branch1:
	la $t1, a
	lw $t4, i
	sll $s2, $t4, 2	# iterator *= 4
	lw $t3, a($s2)		# variable <- array
	sub $t2, $t4, 1		# v1 -> $t2
	la $t1, sum
	sw $t3, v		# spilled v, freed $t3
	sll $s2, $t2, 2	# iterator *= 4
	lw $t3, sum($s2)		# variable <- array
	lw $t1, v
	add $t1, $t1, $t3	# v -> $t1
	sw $t1, v		# spilled v, freed $t1
	la $t1, sum
	sw $t3, v2		# spilled v2, freed $t3
	lw $t3, v
	sll $s2, $t4, 2	# iterator *= 4
	sw $t3, sum($s2)		# variable -> array
	addi $t4, $t4, 1		# i -> $t4
	# Store variables back into memory
	sw $t4, i
	sw $t3, v
	sw $t2, v1

	j loop2

exit2:
	la $t1, sum
	lw $t4, 0($t1)		# variable <- array
	li $t3, 1		# i -> $t3
	# Store variables back into memory
	sw $t4, maxsum
	sw $t3, i

loop3:

	lw $t1, i
	bge $t1, 10, exit3		# exit3 -> $t0
	la $t4, sum
	sll $s2, $t1, 2	# iterator *= 4
	lw $t3, sum($s2)		# variable <- array
	# Store variables back into memory
	sw $t1, i
	sw $t3, v

	lw $t1, maxsum
	lw $t4, v
	bge $t1, $t4, branch2		# branch2 -> $t0
	move $t1, $t4		# maxsum -> $t1
	lw $t3, i
	addi $t3, $t3, 1		# i -> $t3
	# Store variables back into memory
	sw $t1, maxsum
	sw $t4, v
	sw $t3, i

	j loop3

branch2:
	lw $t1, i
	addi $t1, $t1, 1		# i -> $t1
	# Store variables back into memory
	sw $t1, i

	j loop3

exit3:
	li $v0, 1
	lw $t1, maxsum
	move $a0, $t1
	syscall
	# Store variables back into memory
	sw $t1, maxsum
	li $v0, 10
	syscall
	.end main
