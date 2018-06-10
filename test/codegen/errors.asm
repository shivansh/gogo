	.data
a:		.word	0
c:		.word	0
t0:		.word	0
t1:		.word	0
return.0:	.word	0
return.1:	.word	0
a.0:		.word	0
b.1:		.word	0
c.2:		.word	0
d.3:		.word	0
test.0:		.word	0
test.1:		.word	0
e.4:		.word	0
f.5:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime
test:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	lw	$3, test.0	# test.0 -> $3
	move	$5, $3		# a -> $5
	lw	$3, test.1	# test.1 -> $3
	move	$6, $3		# c -> $6
	addi	$3, $5, 1
	addi	$7, $6, 1
	move	$8, $3		# return.0 -> $8
	sw	$8, return.0	# spilled return.0, freed $8
	move	$8, $7		# return.1 -> $8
	# Store dirty variables back into memory
	sw	$3, t0
	sw	$5, a
	sw	$6, c
	sw	$7, t1
	sw	$8, return.1

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end test

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	li	$3, 4		# a.0 -> $3
	sw	$3, a.0		# spilled a.0, freed $3
	li	$3, 1		# b.1 -> $3
	sw	$3, b.1		# spilled b.1, freed $3
	li	$3, 2		# c.2 -> $3
	sw	$3, c.2		# spilled c.2, freed $3
	li	$3, 3		# d.3 -> $3
	sw	$3, d.3		# spilled d.3, freed $3
	li	$3, 1		# test.0 -> $3
	sw	$3, test.0	# spilled test.0, freed $3
	li	$3, 2		# test.1 -> $3
	sw	$3, test.1
	jal	test
	lw	$3, test.1
	lw	$3, return.0	# return.0 -> $3
	move	$5, $3		# e.4 -> $5
	lw	$3, return.1	# return.1 -> $3
	sw	$5, e.4		# spilled e.4, freed $5
	move	$5, $3		# f.5 -> $5
	# Store dirty variables back into memory
	sw	$5, f.5
	li	$2, 10
	syscall
	.end main
