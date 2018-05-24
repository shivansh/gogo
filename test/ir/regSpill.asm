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
	li	$7, 2		# v2 -> $7
	li	$3, -12		# v1 -> $3
	li	$15, 3		# v3 -> $15
	li	$16, 4		# v4 -> $16
	sw	$3, v1		# spilled v1, freed $3
	li	$3, 5		# v5 -> $3
	move	$3, $16		# v5 -> $3
	add	$3, $16, $15	# v5 -> $3
	sw	$7, v2		# spilled v2, freed $7
	li	$7, 5		# v6 -> $7
	li	$8, 5		# v7 -> $8
	sw	$3, v5
	sw	$7, v6
	sw	$8, v7
	sw	$15, v3
	sw	$16, v4
	jal	temp
	lw	$3, v5
	lw	$7, v6
	lw	$8, v7
	lw	$15, v3
	lw	$16, v4
	# Store variables back into memory
	sw	$3, v5
	sw	$7, v6
	sw	$8, v7
	sw	$15, v3
	sw	$16, v4
	li	$v0, 10
	syscall
	.end main
	.globl temp
	.ent temp
temp:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# v1 -> $3
	# Store variables back into memory
	sw	$3, v1

unsat:
	li	$3, 4		# v4 -> $3
	lw	$v0, v1
	# Store variables back into memory
	sw	$3, v4

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end temp
