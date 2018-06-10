	.data
a.0:		.space	12
t0:		.word	0
t1:		.word	0
t2:		.word	0
t3:		.word	0
x.1:		.word	0
t4:		.word	0
y.2:		.word	0
t5:		.word	0
z.3:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime

	.globl main
	.ent main
main:
	la	$3, a.0
	lw	$5, 0($3)	# variable <- array
	sw	$5, t0		# spilled t0, freed $5
	lw	$5, 4($3)	# variable <- array
	sw	$5, t1		# spilled t1, freed $5
	lw	$5, 8($3)	# variable <- array
	sw	$5, t2		# spilled t2, freed $5
	li	$5, 0		# t0 -> $5
	sw	$5, 0($3)	# variable -> array
	li	$6, 1		# t1 -> $6
	sw	$6, 4($3)	# variable -> array
	li	$7, 2		# t2 -> $7
	sw	$7, 8($3)	# variable -> array
	lw	$8, 0($3)	# variable <- array
	move	$9, $8		# x.1 -> $9
	lw	$10, 4($3)	# variable <- array
	move	$11, $10	# y.2 -> $11
	lw	$12, 8($3)	# variable <- array
	move	$13, $12	# z.3 -> $13
	li	$2, 1
	move	$4, $9
	syscall
	li	$2, 1
	move	$4, $11
	syscall
	li	$2, 1
	move	$4, $13
	syscall
	# Store dirty variables back into memory
	sw	$5, t0
	sw	$6, t1
	sw	$7, t2
	sw	$8, t3
	sw	$9, x.1
	sw	$10, t4
	sw	$11, y.2
	sw	$12, t5
	sw	$13, z.3
	li	$2, 10
	syscall
	.end main
