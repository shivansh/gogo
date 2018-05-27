# Test to demonstrate register spilling via next-use heuristic.

	.data
v1:		.word	0
v2:		.word	0
v3:		.word	0
v4:		.word	0
v5:		.word	0
v6:		.word	0
v7:		.word	0

	.text


	.globl main
	.ent main
main:
	li	$3, -1		# v1 -> $3
	sw	$3, v1		# spilled v1, freed $3
	li	$3, 2		# v2 -> $3
	sw	$3, v2		# spilled v2, freed $3
	li	$3, -12		# v1 -> $3
	sw	$3, v1		# spilled v1, freed $3
	li	$3, 3		# v3 -> $3
	li	$5, 4		# v4 -> $5
	li	$6, 5		# v5 -> $6
	move	$6, $5		# v5 -> $6
	add	$6, $5, $3	# v5 -> $6
	sw	$6, v5		# spilled v5, freed $6
	li	$6, 5		# v6 -> $6
	sw	$6, v6		# spilled v6, freed $6
	li	$6, 5		# v7 -> $6
	sw	$3, v3
	sw	$5, v4
	sw	$6, v7
	jal	temp
	lw	$3, v3
	lw	$5, v4
	lw	$6, v7
	# Store dirty variables back into memory
	li	$2, 10
	syscall
	.end main

	.globl temp
	.ent temp
temp:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# v1 -> $3
	# Store dirty variables back into memory
	sw	$3, v1

unsat:
	li	$3, 4		# v4 -> $3
	lw	$2, v1
	# Store dirty variables back into memory
	sw	$3, v4

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end temp
