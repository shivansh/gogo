	.data
a.0:		.word	0
b.1:		.word	0
c.2:		.word	0
t0:		.word	0
t1:		.word	0
t2:		.word	0

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
	li	$3, 1		# a.0 -> $3
	li	$5, 1		# b.1 -> $5
	li	$6, 4		# c.2 -> $6
	sw	$6, c.2		# spilled c.2, freed $6
	add	$6, $3, $5
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, b.1
	sw	$6, t0
	ble	$6, 3, l0

	li	$3, 1		# t1 -> $3
	# Store dirty variables back into memory
	sw	$3, t1
	j	l1

l0:
	li	$3, 0		# t1 -> $3
	# Store dirty variables back into memory
	sw	$3, t1

l1:
	lw	$3, t1		# t1 -> $3
	blt	$3, 1, l2

	li	$2, 1
	lw	$3, a.0		# a.0 -> $3
	move	$4, $3
	syscall
	j	l6

	li	$3, 1		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2
	j	l3

l2:
	li	$3, 0		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2

l3:
	lw	$3, t2		# t2 -> $3
	blt	$3, 1, l5

	li	$2, 1
	lw	$3, b.1		# b.1 -> $3
	move	$4, $3
	syscall
	j	l4

l5:
	li	$2, 1
	lw	$3, c.2		# c.2 -> $3
	move	$4, $3
	syscall

l4:

l6:
	li	$2, 10
	syscall
	.end main
