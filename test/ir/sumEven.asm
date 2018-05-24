# Test to find sum of all even numbers less than n

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
k:		.word	0
l:		.word	0
str:		.asciiz "Sum of all even numbers less than n: "

	.text

	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, nStr
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	li	$7, 0		# i -> $7
	li	$15, 0		# k -> $15
	# Store variables back into memory
	sw	$3, n
	sw	$7, i
	sw	$15, k

loop:
	lw	$3, i
	lw	$7, n
	# Store variables back into memory
	sw	$3, i
	sw	$7, n
	bge	$3, $7, exit	# exit -> $0

	lw	$3, i
	rem	$7, $3, 2	# l -> $7
	# Store variables back into memory
	sw	$3, i
	sw	$7, l
	beq	$7, 1, skip	# skip -> $0

	lw	$3, k
	lw	$7, i
	add	$3, $3, $7	# k -> $3
	addi	$7, $7, 1	# i -> $7
	# Store variables back into memory
	sw	$3, k
	sw	$7, i
	j	loop

skip:
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	j	loop

exit:
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 1
	lw	$3, k
	move	$a0, $3
	syscall
	# Store variables back into memory
	sw	$3, k
	li	$v0, 10
	syscall
	.end main
